package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type AttemptEngineUsecase struct {
	attempts *repository.ExamAttemptRepository
	answers  *repository.ExamAnswerRepository
	students *repository.StudentRepository
	sections *repository.ExamSectionRepository
	exams    *repository.ExamRepository
}

func NewAttemptEngineUsecase(
	attempts *repository.ExamAttemptRepository,
	answers *repository.ExamAnswerRepository,
	students *repository.StudentRepository,
	sections *repository.ExamSectionRepository,
	exams *repository.ExamRepository,
) *AttemptEngineUsecase {
	return &AttemptEngineUsecase{
		attempts: attempts, answers: answers, students: students,
		sections: sections, exams: exams,
	}
}

// resolveAttempt memastikan attempt ada & milik siswa yang login.
func (u *AttemptEngineUsecase) resolveAttempt(ctx context.Context, userID, attemptID uuid.UUID) (*model.ExamAttempt, *model.Exam, error) {
	attempt, err := u.attempts.FindByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperror.NotFound("Attempt tidak ditemukan", err)
		}
		return nil, nil, apperror.Internal(err)
	}

	student, err := u.students.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperror.New(403, "Akun ini bukan siswa", nil)
		}
		return nil, nil, apperror.Internal(err)
	}
	if attempt.StudentID != student.ID {
		return nil, nil, apperror.New(403, "Attempt ini bukan milik Anda", nil)
	}

	exam, err := u.exams.FindByID(ctx, attempt.ExamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, nil, apperror.Internal(err)
	}
	return attempt, exam, nil
}

// expireIfPast menandai attempt expired bila waktu sudah lewat.
func (u *AttemptEngineUsecase) expireIfPast(attempt *model.ExamAttempt) {
	if attempt.Status == model.AttemptStatusInProgress && time.Now().After(attempt.ExpiresAt) {
		_ = u.attempts.MarkExpired(context.Background(), attempt.ID)
		attempt.Status = model.AttemptStatusExpired
	}
}

func (u *AttemptEngineUsecase) ensureActive(attempt *model.ExamAttempt) error {
	u.expireIfPast(attempt)
	if attempt.Status != model.AttemptStatusInProgress {
		return apperror.New(403, "Attempt sudah tidak aktif", nil)
	}
	if time.Now().After(attempt.ExpiresAt) {
		_ = u.attempts.MarkExpired(context.Background(), attempt.ID)
		return apperror.New(403, "Waktu pengerjaan habis", nil)
	}
	return nil
}

func (u *AttemptEngineUsecase) questionInExam(ctx context.Context, exam *model.Exam, questionID uuid.UUID) error {
	exists, err := u.sections.QuestionInExam(ctx, exam.ID, questionID)
	if err != nil {
		return apperror.Internal(err)
	}
	if exists {
		return nil
	}
	// fallback: soal langsung dari bank ujian
	var count int64
	if err := u.exams.CountBankQuestion(ctx, *exam.QuestionBankID, questionID, &count); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// exam tanpa bank → count 0
		_ = err
	}
	if count > 0 {
		return nil
	}
	return apperror.New(422, "Soal bukan bagian dari ujian ini", nil)
}

type AttemptOption struct {
	OptionKey string       `json:"option_key"`
	Text      string       `json:"text"`
	Media     *model.Media `json:"media,omitempty"`
}

type AttemptQuestionItem struct {
	QuestionID    uuid.UUID       `json:"question_id"`
	SectionName   string          `json:"section_name"`
	Sequence      int             `json:"sequence"`
	Type          string          `json:"type"`
	Text          string          `json:"text"`
	ScoreWeight   float64         `json:"score_weight"`
	Media         *model.Media    `json:"media,omitempty"`
	MediaPosition string          `json:"media_position"`
	Options       []AttemptOption `json:"options"`
	AnswerValue   string          `json:"answer_value"`
	IsFlagged     bool            `json:"is_flagged"`
	AnsweredAt    *time.Time      `json:"answered_at"`
}

// GetQuestions mengembalikan lembar soal: daftar soal + jawaban tersimpan + flag.
func (u *AttemptEngineUsecase) GetQuestions(ctx context.Context, userID, attemptID uuid.UUID) (*model.ExamAttempt, []AttemptQuestionItem, error) {
	attempt, exam, err := u.resolveAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, nil, err
	}
	u.expireIfPast(attempt)

	groups, err := u.sections.ListExamQuestions(ctx, exam)
	if err != nil {
		return nil, nil, apperror.Internal(err)
	}

	answers, err := u.answers.ListByAttempt(ctx, attemptID)
	if err != nil {
		return nil, nil, apperror.Internal(err)
	}
	answerByQ := make(map[uuid.UUID]model.ExamAnswer, len(answers))
	for _, a := range answers {
		answerByQ[a.QuestionID] = a
	}

	items := make([]AttemptQuestionItem, 0)
	for _, g := range groups {
		for _, q := range g.Questions {
			item := AttemptQuestionItem{
				QuestionID:    q.ID,
				SectionName:   g.Section.Name,
				Sequence:      g.Section.Sequence,
				Type:          strings.ToUpper(q.QuestionType),
				Text:          q.Content,
				ScoreWeight:   q.ScoreWeight,
				Media:         q.Media,
				MediaPosition: q.MediaPosition,
				Options:       make([]AttemptOption, 0, len(q.Options)),
			}
			for _, o := range q.Options {
				item.Options = append(item.Options, AttemptOption{
					OptionKey: o.Label,
					Text:      o.Content,
					Media:     o.Media,
				})
			}
			if a, ok := answerByQ[q.ID]; ok {
				item.AnswerValue = a.AnswerValue
				item.IsFlagged = a.IsFlagged
				at := a.AnsweredAt
				item.AnsweredAt = &at
			}
			items = append(items, item)
		}
	}
	return attempt, items, nil
}

type SaveAnswerInput struct {
	QuestionID      uuid.UUID
	AnswerValue     string
	ClientTimestamp int64
}

// SaveAnswer menyimpan jawaban real-time (upsert) selama attempt aktif.
func (u *AttemptEngineUsecase) SaveAnswer(ctx context.Context, userID, attemptID uuid.UUID, in SaveAnswerInput) (*model.ExamAnswer, error) {
	attempt, exam, err := u.resolveAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	if err := u.ensureActive(attempt); err != nil {
		return nil, err
	}
	if err := u.questionInExam(ctx, exam, in.QuestionID); err != nil {
		return nil, err
	}
	return u.answers.UpsertAnswer(ctx, attemptID, in.QuestionID, in.AnswerValue, in.ClientTimestamp)
}

// SetFlag menandai/melepas ragu-ragu pada soal.
func (u *AttemptEngineUsecase) SetFlag(ctx context.Context, userID, attemptID, questionID uuid.UUID, flagged bool) (*model.ExamAnswer, error) {
	attempt, exam, err := u.resolveAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	if err := u.ensureActive(attempt); err != nil {
		return nil, err
	}
	if err := u.questionInExam(ctx, exam, questionID); err != nil {
		return nil, err
	}
	return u.answers.SetFlag(ctx, attemptID, questionID, flagged)
}

type HeartbeatResult struct {
	ServerTime       time.Time `json:"server_time"`
	RemainingSeconds int64     `json:"remaining_seconds"`
	IsExpired        bool      `json:"is_expired"`
}

// Heartbeat: sumber kebenaran sisa waktu adalah server, bukan jam browser.
func (u *AttemptEngineUsecase) Heartbeat(ctx context.Context, userID, attemptID uuid.UUID) (*model.ExamAttempt, HeartbeatResult, error) {
	attempt, _, err := u.resolveAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, HeartbeatResult{}, err
	}
	u.expireIfPast(attempt)

	now := time.Now()
	remaining := attempt.ExpiresAt.Sub(now).Milliseconds() / 1000
	if remaining < 0 {
		remaining = 0
	}
	isExpired := attempt.Status != model.AttemptStatusInProgress || remaining == 0

	return attempt, HeartbeatResult{
		ServerTime:       now,
		RemainingSeconds: remaining,
		IsExpired:        isExpired,
	}, nil
}

type AutosaveItem struct {
	QuestionID uuid.UUID
	Value      string
}

// Autosave menyimpan banyak jawaban sekaligus (batch interval FE).
func (u *AttemptEngineUsecase) Autosave(ctx context.Context, userID, attemptID uuid.UUID, items []AutosaveItem) (int, error) {
	attempt, exam, err := u.resolveAttempt(ctx, userID, attemptID)
	if err != nil {
		return 0, err
	}
	if err := u.ensureActive(attempt); err != nil {
		return 0, err
	}

	saved := 0
	for _, item := range items {
		if err := u.questionInExam(ctx, exam, item.QuestionID); err != nil {
			continue
		}
		if _, err := u.answers.UpsertAnswer(ctx, attemptID, item.QuestionID, item.Value, time.Now().Unix()); err != nil {
			return saved, err
		}
		saved++
	}
	return saved, nil
}

// Submit finalisasi attempt: status submitted + submitted_at.
func (u *AttemptEngineUsecase) Submit(ctx context.Context, userID, attemptID uuid.UUID, confirm bool) (*model.ExamAttempt, error) {
	attempt, _, err := u.resolveAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	if !confirm {
		return nil, apperror.New(422, "Konfirmasi pengumpulan diperlukan", nil)
	}
	if attempt.Status == model.AttemptStatusSubmitted {
		return attempt, nil
	}

	now := time.Now()
	if attempt.Status == model.AttemptStatusInProgress && now.After(attempt.ExpiresAt) {
		_ = u.attempts.MarkExpired(ctx, attempt.ID)
	}

	attempt.Status = model.AttemptStatusSubmitted
	attempt.SubmittedAt = &now

	if err := u.attempts.FinalizeSubmit(ctx, attempt.ID, now); err != nil {
		return nil, err
	}
	return u.repo_findByID(ctx, attemptID)
}

func (u *AttemptEngineUsecase) repo_findByID(ctx context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
	return u.attempts.FindByID(ctx, id)
}
