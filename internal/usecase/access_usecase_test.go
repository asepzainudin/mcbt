package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

func ownedBank(owner uuid.UUID) *model.QuestionBank {
	return &model.QuestionBank{BaseModel: model.BaseModel{ID: uuid.New()}, CreatedBy: &owner}
}

func newAccessUC() (*AccessUsecase, *fakeBankRepo, *fakeExamRepo, *fakeSectionRepo, *fakeQuestionRepo, *fakeAttemptRepo) {
	banks := &fakeBankRepo{}
	exams := &fakeExamRepo{}
	sections := &fakeSectionRepo{}
	questions := &fakeQuestionRepo{}
	attempts := &fakeAttemptRepo{}
	uc := NewAccessUsecase(banks, exams, sections, questions, attempts)
	return uc, banks, exams, sections, questions, attempts
}

func TestAccessAssertBankOwner(t *testing.T) {
	owner := uuid.New()
	uc, banks, _, _, _, _ := newAccessUC()
	bank := ownedBank(owner)
	banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) { return bank, nil }

	if err := uc.AssertBankOwner(ctxBg(), owner, false, bank.ID); err != nil {
		t.Errorf("pemilik harus lolos: %v", err)
	}
	err := uc.AssertBankOwner(ctxBg(), uuid.New(), false, bank.ID)
	if ae := apperror.From(err); ae.Code != 403 {
		t.Errorf("non-pemilik harus 403, got %v", err)
	}
	if err := uc.AssertBankOwner(ctxBg(), uuid.New(), true, bank.ID); err != nil {
		t.Errorf("admin harus selalu lolos: %v", err)
	}

	// bank tanpa creator (data lama): guru ditolak
	orphan := &model.QuestionBank{BaseModel: model.BaseModel{ID: uuid.New()}}
	banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) { return orphan, nil }
	if err := uc.AssertBankOwner(ctxBg(), owner, false, orphan.ID); err == nil {
		t.Error("bank tanpa pemilik harus ditolak utk guru")
	}
}

func TestAccessAssertExamOwner(t *testing.T) {
	owner := uuid.New()
	uc, _, exams, _, _, _ := newAccessUC()

	t.Run("ujian milik pembuat", func(t *testing.T) {
		exam := &model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}, CreatedBy: &owner}
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return exam, nil }
		if err := uc.AssertExamOwner(ctxBg(), owner, false, exam.ID); err != nil {
			t.Errorf("pembuat harus lolos: %v", err)
		}
	})

	t.Run("ujian orang lain 403 / admin lolos", func(t *testing.T) {
		exam := &model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}, CreatedBy: &owner}
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return exam, nil }
		err := uc.AssertExamOwner(ctxBg(), uuid.New(), false, exam.ID)
		if ae := apperror.From(err); ae.Code != 403 {
			t.Errorf("want 403, got %v", err)
		}
		if err := uc.AssertExamOwner(ctxBg(), uuid.Nil, true, exam.ID); err != nil {
			t.Errorf("admin harus lolos: %v", err)
		}
	})

	t.Run("ujian tanpa created_by ditolak utk guru", func(t *testing.T) {
		exam := &model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}}
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return exam, nil }
		if err := uc.AssertExamOwner(ctxBg(), owner, false, exam.ID); err == nil {
			t.Error("created_by NULL harus ditolak")
		}
	})
}

func TestAccessAssertChain(t *testing.T) {
	owner := uuid.New()
	stranger := uuid.New()
	uc, _, exams, sections, questions, attempts := newAccessUC()

	exam := &model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}, CreatedBy: &owner}
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return exam, nil }

	section := &model.ExamSection{BaseModel: model.BaseModel{ID: uuid.New()}, ExamID: exam.ID}
	sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) { return section, nil }

	question := &model.Question{BaseModel: model.BaseModel{ID: uuid.New()}, QuestionBankID: uuid.New()}
	questions.findByIDFn = func(context.Context, uuid.UUID) (*model.Question, error) { return question, nil }

	attempt := &model.ExamAttempt{BaseModel: model.BaseModel{ID: uuid.New()}, ExamID: exam.ID}
	attempts.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamAttempt, error) { return attempt, nil }

	// section → exam: pemilik lolos
	if err := uc.AssertSectionOwner(ctxBg(), owner, false, section.ID); err != nil {
		t.Errorf("section chain gagal utk pemilik: %v", err)
	}
	// attempt → exam: orang lain ditolak
	err := uc.AssertAttemptOwner(ctxBg(), stranger, false, attempt.ID)
	if ae := apperror.From(err); ae.Code != 403 {
		t.Errorf("attempt asing want 403, got %v", err)
	}
	// question → bank: bank tak dikenal → 404
	q2 := &model.Question{BaseModel: model.BaseModel{ID: uuid.New()}, QuestionBankID: uuid.New()}
	questions.findByIDFn = func(context.Context, uuid.UUID) (*model.Question, error) { return q2, nil }
	errQ := uc.AssertQuestionOwner(ctxBg(), stranger, false, q2.ID)
	if ae := apperror.From(errQ); ae.Code != 404 {
		t.Errorf("question→bank not found want 404, got %v", errQ)
	}
}

func TestAccessNotFound(t *testing.T) {
	uc, _, _, _, _, _ := newAccessUC()
	// semua default fake → ErrRecordNotFound
	err := uc.AssertExamOwner(ctxBg(), uuid.New(), false, uuid.New())
	if ae := apperror.From(err); ae.Code != 404 {
		t.Errorf("ujian tak ada want 404, got %v", err)
	}
	err = uc.AssertQuestionOwner(ctxBg(), uuid.New(), false, uuid.New())
	if ae := apperror.From(err); ae.Code != 404 {
		t.Errorf("soal tak ada want 404, got %v", err)
	}
}
