package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type OptionInput struct {
	Content   string
	MediaID   *uuid.UUID
	IsCorrect bool
}

type QuestionInput struct {
	BankID        uuid.UUID
	Type          string
	Content       string
	ScoreWeight   float64
	Explanation   *string
	MediaID       *uuid.UUID
	MediaPosition string
	Options       []OptionInput
	AnswerKeys    []string
}

type QuestionUsecase struct {
	repo     *repository.QuestionRepository
	banks    *repository.QuestionBankRepository
	sections *repository.ExamSectionRepository
	answers  *repository.ExamAnswerRepository
}

func NewQuestionUsecase(
	repo *repository.QuestionRepository,
	banks *repository.QuestionBankRepository,
	sections *repository.ExamSectionRepository,
	answers *repository.ExamAnswerRepository,
) *QuestionUsecase {
	return &QuestionUsecase{repo: repo, banks: banks, sections: sections, answers: answers}
}

var errQuestionInUse = errors.New("soal sudah digunakan ujian")

// ensureNotUsed menolak perubahan/penghapusan soal yang sudah digunakan ujian
// (termapping di section maupun sudah dijawab siswa).
func (u *QuestionUsecase) ensureNotUsed(ctx context.Context, questionID uuid.UUID) error {
	mapped, err := u.sections.UsedQuestionIDs(ctx, []uuid.UUID{questionID})
	if err != nil {
		return apperror.Internal(err)
	}
	if mapped[questionID] {
		return apperror.New(409, "Soal sudah digunakan ujian. Mohon buat soal baru!", errQuestionInUse)
	}
	answered, err := u.answers.AnsweredQuestionIDs(ctx, []uuid.UUID{questionID})
	if err != nil {
		return apperror.Internal(err)
	}
	if answered[questionID] {
		return apperror.New(409, "Soal sudah digunakan ujian. Mohon buat soal baru!", errQuestionInUse)
	}
	return nil
}

var validQuestionTypes = map[string]bool{
	model.QuestionTypeMultipleChoice: true,
	model.QuestionTypeTrueFalse:      true,
	model.QuestionTypeMultipleAnswer: true,
	model.QuestionTypeEssay:          true,
	model.QuestionTypeShortAnswer:    true,
}

const (
	maxOptions    = 5
	minOptions    = 2
	maxScoreValue = 999.99
)

func validateType(t string) error {
	if !validQuestionTypes[t] {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"question_type": fmt.Sprintf(
				"tipe harus salah satu dari %s, %s, %s, %s, %s",
				model.QuestionTypeMultipleChoice, model.QuestionTypeTrueFalse,
				model.QuestionTypeMultipleAnswer, model.QuestionTypeEssay, model.QuestionTypeShortAnswer,
			)},
		}
	}
	return nil
}

func normalizeMediaPosition(pos string) string {
	if pos == "before" {
		return "before"
	}
	return "after"
}

func validateScore(w float64) float64 {
	if w <= 0 {
		return 1.0
	}
	if w > maxScoreValue {
		return maxScoreValue
	}
	return w
}

// buildOptions normalises options per question type, assigning labels/positions.
// Returns options and whether at least one correct answer is marked.
func buildOptions(qType string, inputs []OptionInput) ([]model.Option, bool, error) {
	if qType == model.QuestionTypeEssay || qType == model.QuestionTypeShortAnswer {
		return nil, false, nil
	}

	if qType == model.QuestionTypeTrueFalse && len(inputs) == 0 {
		inputs = []OptionInput{{Content: "BENAR"}, {Content: "SALAH"}}
	}

	n := len(inputs)
	if n < minOptions || n > maxOptions {
		return nil, false, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"options": fmt.Sprintf(
				"soal ini membutuhkan %d–%d opsi", minOptions, maxOptions,
			)},
		}
	}

	correctCount := 0
	for _, o := range inputs {
		if o.Content == "" && o.MediaID == nil {
			return nil, false, &apperror.AppError{
				Code:    apperror.CodeUnprocessable,
				Message: "Validasi gagal",
				Details: map[string]string{"options": "setiap opsi wajib memiliki teks atau gambar"},
			}
		}
		if o.IsCorrect {
			correctCount++
		}
	}

	if qType == model.QuestionTypeMultipleChoice && correctCount != 1 {
		return nil, false, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"options": "pilihan ganda harus memiliki tepat satu jawaban benar"},
		}
	}
	if qType == model.QuestionTypeTrueFalse && correctCount != 1 {
		return nil, false, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"options": "benar/salah harus memiliki tepat satu jawaban benar"},
		}
	}
	if qType == model.QuestionTypeMultipleAnswer && correctCount < 1 {
		return nil, false, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"options": "multiple answer minimal satu jawaban benar"},
		}
	}

	opts := make([]model.Option, n)
	for i, o := range inputs {
		opts[i] = model.Option{
			Label:     string(rune('A' + i)),
			Content:   o.Content,
			MediaID:   o.MediaID,
			IsCorrect: o.IsCorrect,
			Position:  int16(i),
		}
	}
	return opts, true, nil
}

func validateAnswerKeys(keys []string) error {
	if len(keys) == 0 {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"answer_keys": "isian singkat wajib memiliki minimal satu jawaban yang diterima"},
		}
	}
	return nil
}

func (u *QuestionUsecase) bankExists(ctx context.Context, id uuid.UUID) error {
	_, err := u.banks.FindByID(ctx, id)
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"question_bank_id": "bank soal tidak ditemukan"},
		}
	}
	return apperror.Internal(err)
}

func (u *QuestionUsecase) List(ctx context.Context, p repository.QuestionListParams) ([]model.Question, int64, error) {
	result, err := u.repo.List(ctx, p)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}

	ids := make([]uuid.UUID, 0, len(result.Items))
	for _, q := range result.Items {
		ids = append(ids, q.ID)
	}
	mappedSet, err := u.sections.UsedQuestionIDs(ctx, ids)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	answeredSet, err := u.answers.AnsweredQuestionIDs(ctx, ids)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	for i := range result.Items {
		result.Items[i].IsUsed = mappedSet[result.Items[i].ID] || answeredSet[result.Items[i].ID]
	}
	return result.Items, result.TotalItems, nil
}

func (u *QuestionUsecase) Get(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	q, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	return q, nil
}

func (u *QuestionUsecase) Create(ctx context.Context, in QuestionInput) (*model.Question, error) {
	if err := u.bankExists(ctx, in.BankID); err != nil {
		return nil, err
	}
	if err := validateType(in.Type); err != nil {
		return nil, err
	}
	if in.Type == model.QuestionTypeShortAnswer {
		if err := validateAnswerKeys(in.AnswerKeys); err != nil {
			return nil, err
		}
	}

	opts, _, err := buildOptions(in.Type, in.Options)
	if err != nil {
		return nil, err
	}

	question := &model.Question{
		QuestionBankID: in.BankID,
		MediaID:        in.MediaID,
		QuestionType:   in.Type,
		Content:        in.Content,
		ScoreWeight:    validateScore(in.ScoreWeight),
		Explanation:    in.Explanation,
		AnswerKeys:     joinKeys(in.AnswerKeys),
		MediaPosition:  normalizeMediaPosition(in.MediaPosition),
		Options:        opts,
	}
	if err := u.repo.CreateWithOptions(ctx, question); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, question.ID)
}

func (u *QuestionUsecase) Update(ctx context.Context, id uuid.UUID, in QuestionInput) (*model.Question, error) {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	if err := u.ensureNotUsed(ctx, id); err != nil {
		return nil, err
	}

	// Spec PUT tidak mengirim bank_id: pertahankan bank asal bila tidak dikirim.
	if in.BankID == uuid.Nil {
		in.BankID = existing.QuestionBankID
	} else if in.BankID != existing.QuestionBankID {
		if err := u.bankExists(ctx, in.BankID); err != nil {
			return nil, err
		}
	}
	if err := validateType(in.Type); err != nil {
		return nil, err
	}

	var opts []model.Option
	if in.Type != model.QuestionTypeEssay && in.Type != model.QuestionTypeShortAnswer {
		var buildErr error
		opts, _, buildErr = buildOptions(in.Type, in.Options)
		if buildErr != nil {
			return nil, buildErr
		}
	} else {
		opts = nil
	}

	if in.Type == model.QuestionTypeShortAnswer {
		if err := validateAnswerKeys(in.AnswerKeys); err != nil {
			return nil, err
		}
	}

	existing.QuestionBankID = in.BankID
	existing.MediaID = in.MediaID
	existing.QuestionType = in.Type
	existing.Content = in.Content
	existing.ScoreWeight = validateScore(in.ScoreWeight)
	existing.Explanation = in.Explanation
	existing.AnswerKeys = joinKeys(in.AnswerKeys)
	existing.MediaPosition = normalizeMediaPosition(in.MediaPosition)

	if err := u.repo.UpdateWithOptions(ctx, existing, opts); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *QuestionUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := u.ensureNotUsed(ctx, id); err != nil {
		return err
	}
	err := u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Soal tidak ditemukan", err)
		}
		return err
	}
	return nil
}

func (u *QuestionUsecase) ReorderOptions(ctx context.Context, questionID uuid.UUID, orderedIDs []uuid.UUID) error {
	if _, err := u.Get(ctx, questionID); err != nil {
		return err
	}
	return u.repo.ReorderOptions(ctx, questionID, orderedIDs)
}

func (u *QuestionUsecase) SetCorrectOption(ctx context.Context, questionID, optionID uuid.UUID) (*model.Question, error) {
	if _, err := u.Get(ctx, questionID); err != nil {
		return nil, err
	}
	if err := u.repo.SetCorrectOption(ctx, questionID, optionID); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, questionID)
}

func joinKeys(keys []string) *string {
	cleaned := make([]string, 0, len(keys))
	for _, k := range keys {
		k = trimSpace(k)
		if k != "" {
			cleaned = append(cleaned, k)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	joined := cleaned[0]
	for _, k := range cleaned[1:] {
		joined += "\n" + k
	}
	return &joined
}
