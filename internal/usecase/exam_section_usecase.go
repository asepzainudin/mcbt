package usecase

import (
	"context"
	"errors"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type ExamSectionUsecase struct {
	sections *repository.ExamSectionRepository
	exams    *repository.ExamRepository
	banks    *repository.QuestionBankRepository
}

func NewExamSectionUsecase(
	sections *repository.ExamSectionRepository,
	exams *repository.ExamRepository,
	banks *repository.QuestionBankRepository,
) *ExamSectionUsecase {
	return &ExamSectionUsecase{sections: sections, exams: exams, banks: banks}
}

type ExamSectionInput struct {
	Name     string
	Sequence int
}

type MapQuestionsInput struct {
	BankIDs              []uuid.UUID
	TotalRandomQuestions int
}

func (u *ExamSectionUsecase) ListByExam(ctx context.Context, examID uuid.UUID) ([]repository.SectionWithCount, error) {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	sections, err := u.sections.ListByExam(ctx, examID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return sections, nil
}

func (u *ExamSectionUsecase) validateInput(ctx context.Context, examID uuid.UUID, in ExamSectionInput) error {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	if strings.TrimSpace(in.Name) == "" {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"name": "nama section wajib diisi"},
		}
	}
	if in.Sequence < 1 {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"sequence": "sequence minimal 1"},
		}
	}
	return nil
}

func (u *ExamSectionUsecase) Create(ctx context.Context, examID uuid.UUID, in ExamSectionInput) (*model.ExamSection, error) {
	if err := u.validateInput(ctx, examID, in); err != nil {
		return nil, err
	}

	section := &model.ExamSection{
		ExamID:   examID,
		Name:     in.Name,
		Sequence: in.Sequence,
	}
	if err := u.sections.Create(ctx, section); err != nil {
		return nil, err
	}
	return section, nil
}

func (u *ExamSectionUsecase) Update(ctx context.Context, id uuid.UUID, in ExamSectionInput) (*model.ExamSection, error) {
	section, err := u.sections.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Section tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	if err := u.validateInput(ctx, section.ExamID, in); err != nil {
		return nil, err
	}

	section.Name = in.Name
	section.Sequence = in.Sequence
	if err := u.sections.Update(ctx, section); err != nil {
		return nil, err
	}
	return section, nil
}

func (u *ExamSectionUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.sections.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Section tidak ditemukan", err)
		}
		return err
	}
	return nil
}

// MapQuestions snapshots questions from the given banks into the section.
// When TotalRandomQuestions > 0, a random sample of that size is taken from
// the union of bank questions (capped at availability).
func (u *ExamSectionUsecase) MapQuestions(ctx context.Context, sectionID uuid.UUID, in MapQuestionsInput) (int, int, error) {
	if len(in.BankIDs) == 0 {
		return 0, 0, apperror.New(422, "Pilih minimal satu bank soal", nil)
	}

	section, err := u.sections.FindByID(ctx, sectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, apperror.NotFound("Section tidak ditemukan", err)
		}
		return 0, 0, apperror.Internal(err)
	}

	for _, bankID := range in.BankIDs {
		if _, err := u.banks.FindByID(ctx, bankID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, 0, &apperror.AppError{
					Code:    apperror.CodeUnprocessable,
					Message: "Validasi gagal",
					Details: map[string]string{"question_bank_ids": "bank soal tidak ditemukan"},
				}
			}
			return 0, 0, apperror.Internal(err)
		}
	}

	available, err := u.sections.QuestionIDsByBanks(ctx, in.BankIDs)
	if err != nil {
		return 0, 0, apperror.Internal(err)
	}
	if len(available) == 0 {
		return 0, 0, apperror.New(422, "Bank soal terpilih tidak memiliki soal", nil)
	}

	mappedSet, err := u.sections.MappedQuestionIDsByExam(ctx, section.ExamID)
	if err != nil {
		return 0, 0, apperror.Internal(err)
	}

	// kandidat: belum termapping di section/exam ini
	candidates := make([]uuid.UUID, 0, len(available))
	for _, id := range available {
		if !mappedSet[id] {
			candidates = append(candidates, id)
		}
	}

	// sampling acak bila total_random_questions ditentukan
	limit := in.TotalRandomQuestions
	if limit > 0 && limit < len(candidates) {
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		candidates = candidates[:limit]
	}

	inserted, err := u.sections.InsertMappings(ctx, sectionID, candidates)
	if err != nil {
		return 0, 0, err
	}

	skipped := len(available) - inserted
	return inserted, skipped, nil
}

func (u *ExamSectionUsecase) RemoveQuestion(ctx context.Context, sectionID, questionID uuid.UUID) error {
	err := u.sections.RemoveQuestion(ctx, sectionID, questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Soal tidak termapping di section ini", err)
		}
		return err
	}
	return nil
}

func (u *ExamSectionUsecase) ListQuestions(ctx context.Context, sectionID uuid.UUID) ([]model.Question, error) {
	if _, err := u.sections.FindByID(ctx, sectionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Section tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	questions, err := u.sections.ListQuestions(ctx, sectionID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return questions, nil
}
