package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	passwordutil "github.com/asepzainudin14/mcbt/internal/pkg/password"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type ExamUsecase struct {
	repo     *repository.ExamRepository
	subjects *repository.SubjectRepository
	ays      *repository.AcademicYearRepository
	banks    *repository.QuestionBankRepository
	attempts *repository.ExamAttemptRepository
}

func NewExamUsecase(
	repo *repository.ExamRepository,
	subjects *repository.SubjectRepository,
	ays *repository.AcademicYearRepository,
	banks *repository.QuestionBankRepository,
	attempts *repository.ExamAttemptRepository,
) *ExamUsecase {
	return &ExamUsecase{repo: repo, subjects: subjects, ays: ays, banks: banks, attempts: attempts}
}

type ExamInput struct {
	Title          string
	Description    *string
	SubjectID      uuid.UUID
	AcademicYearID *uuid.UUID
	QuestionBankID *uuid.UUID
	CreatedBy      *uuid.UUID // pembuat ujian (untuk data scope guru)
}

type ExamSettingsInput struct {
	DurationMinutes       int
	MaxAttempts           int
	PassingGrade          float64
	RandomizeQuestions    bool
	RandomizeOptions      bool
	AllowBacktrack        bool
	AutoSubmit            bool
	ShowResultImmediately bool
	NegativeMarking       bool
	NegativeValue         float64
	TokenEnabled          bool
	AllowDiscussion       bool
}

func refExists(ctx context.Context, err error, field, message string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{field: message},
		}
	}
	return apperror.Internal(err)
}

func (u *ExamUsecase) validateRefs(ctx context.Context, in ExamInput) error {
	if _, err := u.subjects.FindByID(ctx, in.SubjectID); err != nil {
		return refExists(ctx, err, "subject_id", "mapel tidak ditemukan")
	}
	if in.AcademicYearID != nil {
		if _, err := u.ays.FindByID(ctx, *in.AcademicYearID); err != nil {
			return refExists(ctx, err, "academic_year_id", "tahun ajaran tidak ditemukan")
		}
	}
	if in.QuestionBankID != nil {
		if _, err := u.banks.FindByID(ctx, *in.QuestionBankID); err != nil {
			return refExists(ctx, err, "question_bank_id", "bank soal tidak ditemukan")
		}
	}
	return nil
}

func validateSettings(in *ExamSettingsInput) error {
	e := map[string]string{}

	if in.DurationMinutes < 1 || in.DurationMinutes > 600 {
		e["duration_minutes"] = "durasi harus 1–600 menit"
	}
	if in.MaxAttempts < 1 || in.MaxAttempts > 10 {
		e["max_attempts"] = "attempt harus 1–10"
	}
	if in.PassingGrade < 0 || in.PassingGrade > 100 {
		e["passing_grade"] = "passing grade harus 0–100"
	}
	if in.NegativeValue < 0 || in.NegativeValue > 100 {
		e["negative_value"] = "nilai negatif harus 0–100"
	}
	if in.NegativeMarking && in.NegativeValue == 0 {
		in.NegativeValue = 1.0
	}
	if !in.NegativeMarking {
		in.NegativeValue = 0
	}

	if len(e) > 0 {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: e,
		}
	}
	return nil
}

func (u *ExamUsecase) List(ctx context.Context, p repository.ExamListParams) ([]model.Exam, int64, error) {
	result, err := u.repo.List(ctx, p)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	for i := range result.Items {
		count, err := u.attempts.CountByExam(ctx, result.Items[i].ID)
		if err == nil {
			result.Items[i].AttemptsCount = count
		}
	}
	return result.Items, result.TotalItems, nil
}

func (u *ExamUsecase) Get(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	e, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	return e, nil
}

func (u *ExamUsecase) Create(ctx context.Context, in ExamInput) (*model.Exam, error) {
	if err := u.validateRefs(ctx, in); err != nil {
		return nil, err
	}

	exam := &model.Exam{
		Title:          in.Title,
		Description:    in.Description,
		SubjectID:      in.SubjectID,
		AcademicYearID: in.AcademicYearID,
		QuestionBankID: in.QuestionBankID,
		CreatedBy:      in.CreatedBy,
		Status:         model.ExamStatusDraft,
	}
	if err := u.repo.Create(ctx, exam); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, exam.ID)
}

func (u *ExamUsecase) Update(ctx context.Context, id uuid.UUID, in ExamInput) (*model.Exam, error) {
	exam, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	if err := u.validateRefs(ctx, in); err != nil {
		return nil, err
	}

	exam.Title = in.Title
	exam.Description = in.Description
	exam.SubjectID = in.SubjectID
	exam.AcademicYearID = in.AcademicYearID
	exam.QuestionBankID = in.QuestionBankID

	if err := u.repo.UpdateCore(ctx, exam); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *ExamUsecase) UpdateSettings(ctx context.Context, id uuid.UUID, in ExamSettingsInput) (*model.Exam, error) {
	exam, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	if err := validateSettings(&in); err != nil {
		return nil, err
	}

	exam.DurationMinutes = in.DurationMinutes
	exam.MaxAttempts = in.MaxAttempts
	exam.PassingGrade = in.PassingGrade
	exam.RandomizeQuestions = in.RandomizeQuestions
	exam.RandomizeOptions = in.RandomizeOptions
	exam.AllowBacktrack = in.AllowBacktrack
	exam.AutoSubmit = in.AutoSubmit
	exam.ShowResultImmediately = in.ShowResultImmediately
	exam.NegativeMarking = in.NegativeMarking
	exam.NegativeValue = in.NegativeValue
	exam.TokenEnabled = in.TokenEnabled
	exam.AllowDiscussion = in.AllowDiscussion

	if in.TokenEnabled && (exam.ExamToken == nil || *exam.ExamToken == "") {
		token, err := passwordutil.Generate(6)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		exam.ExamToken = &token
	}
	if !in.TokenEnabled {
		exam.ExamToken = nil
	}

	if err := u.repo.UpdateSettings(ctx, exam); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *ExamUsecase) Publish(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	if err := u.repo.SetStatus(ctx, id, model.ExamStatusPublished); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *ExamUsecase) Close(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	if err := u.repo.SetStatus(ctx, id, model.ExamStatusClosed); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *ExamUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	used, err := u.attempts.CountByExam(ctx, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if used > 0 {
		return apperror.New(409, "Ujian sudah digunakan peserta dan tidak dapat dihapus", nil)
	}
	err = u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return err
	}
	return nil
}
