package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

func newBankUC() (*QuestionBankUsecase, *fakeBankRepo, *fakeSubjectRepo, *fakeQuestionRepo) {
	repo := &fakeBankRepo{}
	subjects := &fakeSubjectRepo{}
	questions := &fakeQuestionRepo{}
	return NewQuestionBankUsecase(repo, subjects, questions), repo, subjects, questions
}

func bankOf(owner *uuid.UUID) *model.QuestionBank {
	return &model.QuestionBank{BaseModel: model.BaseModel{ID: uuid.New()}, CreatedBy: owner, Status: model.BankStatusDraft}
}

func TestBankCreate_SetsDraftAndFields(t *testing.T) {
	uc, repo, subjects, _ := newBankUC()
	subjectID := uuid.New()
	subjects.findByIDFn = func(context.Context, uuid.UUID) (*model.Subject, error) {
		return &model.Subject{BaseModel: model.BaseModel{ID: subjectID}}, nil
	}
	var saved *model.QuestionBank
	repo.createFn = func(_ context.Context, qb *model.QuestionBank) error {
		saved = qb
		qb.ID = uuid.New()
		return nil
	}
	repo.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.QuestionBank, error) {
		return saved, nil
	}

	owner := uuid.New()
	out, err := uc.Create(ctxBg(), QuestionBankInput{
		SubjectID: subjectID, Code: "BK1", Title: "Bank 1", CreatedBy: &owner,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if saved.Status != model.BankStatusDraft {
		t.Errorf("status = %q, want draft", saved.Status)
	}
	if saved.CreatedBy == nil || *saved.CreatedBy != owner {
		t.Error("CreatedBy tidak di-set dari input")
	}
	if out == nil || out.Code != "BK1" {
		t.Error("hasil FindByID tidak dikembalikan")
	}
}

func TestBankCreate_SubjectMissing(t *testing.T) {
	uc, _, subjects, _ := newBankUC()
	subjects.findByIDFn = func(context.Context, uuid.UUID) (*model.Subject, error) { return nil, notFound() }

	_, err := uc.Create(ctxBg(), QuestionBankInput{SubjectID: uuid.New()})
	if ae := apperror.From(err); ae.Code != 422 {
		t.Fatalf("want 422, got %v", err)
	}
}

func TestBankPublish_EmptyBankRejected(t *testing.T) {
	uc, repo, _, questions := newBankUC()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
		return bankOf(nil), nil
	}
	questions.countByBank = map[uuid.UUID]int64{} // 0 soal

	_, err := uc.Publish(ctxBg(), uuid.New())
	if ae := apperror.From(err); ae.Code != 422 {
		t.Fatalf("bank kosong want 422, got %v", err)
	}
}

func TestBankPublish_WithQuestions(t *testing.T) {
	uc, repo, _, questions := newBankUC()
	id := uuid.New()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
		return bankOf(nil), nil
	}
	questions.countByBank = map[uuid.UUID]int64{id: 3}

	if _, err := uc.Publish(ctxBg(), id); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if len(repo.setStatus) != 1 || repo.setStatus[0].status != model.BankStatusPublished {
		t.Errorf("SetStatus(published) tidak dipanggil: %+v", repo.setStatus)
	}
}

func TestBankClone_OwnedByActor(t *testing.T) {
	uc, repo, _, _ := newBankUC()
	actor := uuid.New()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
		return bankOf(nil), nil
	}

	clone, err := uc.Clone(ctxBg(), uuid.New(), actor)
	if err != nil {
		t.Fatalf("clone: %v", err)
	}
	if repo.cloneOwner == nil || *repo.cloneOwner != actor {
		t.Error("clone harus dimiliki pelaku, bukan pemilik sumber")
	}
	if clone.ID == repo.cloneSrc.ID {
		t.Error("clone harus punya ID baru")
	}
}

func TestBankDelete_NotFound(t *testing.T) {
	uc, repo, _, _ := newBankUC()
	repo.deleteErr = gorm.ErrRecordNotFound

	if ae := apperror.From(uc.Delete(ctxBg(), uuid.New())).Code; ae != 404 {
		t.Fatalf("want 404, got %v", ae)
	}
}

// ---------- QuestionUsecase ----------

func newQuestionUC() (*QuestionUsecase, *fakeQuestionRepo, *fakeBankRepo, *fakeSectionRepo, *fakeAnswerRepo) {
	repo := &fakeQuestionRepo{}
	banks := &fakeBankRepo{}
	sections := &fakeSectionRepo{}
	answers := &fakeAnswerRepo{}
	uc := NewQuestionUsecase(repo, banks, sections, answers)
	return uc, repo, banks, sections, answers
}

func mcOptions() []OptionInput {
	return []OptionInput{{Content: "satu"}, {Content: "dua", IsCorrect: true}}
}

func TestQuestionCreate_Validations(t *testing.T) {
	t.Run("bank tidak ada → 422/404", func(t *testing.T) {
		uc, _, banks, _, _ := newQuestionUC()
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return nil, notFound()
		}
		_, err := uc.Create(ctxBg(), QuestionInput{BankID: uuid.New(), Type: model.QuestionTypeMultipleChoice, Options: mcOptions()})
		if apperror.From(err).Code == 0 {
			t.Fatalf("harus error, got %v", err)
		}
	})

	t.Run("MC 1 opsi → 422", func(t *testing.T) {
		uc, _, banks, _, _ := newQuestionUC()
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return bankOf(nil), nil
		}
		_, err := uc.Create(ctxBg(), QuestionInput{
			BankID: uuid.New(), Type: model.QuestionTypeMultipleChoice,
			Options: []OptionInput{{Content: "satu"}},
		})
		ae := apperror.From(err)
		if ae.Code != 422 || !strings.Contains(ae.Details["options"], "2") {
			t.Fatalf("want 422 min 2 opsi, got %v", err)
		}
	})

	t.Run("tipe tidak valid → 422", func(t *testing.T) {
		uc, _, banks, _, _ := newQuestionUC()
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return bankOf(nil), nil
		}
		_, err := uc.Create(ctxBg(), QuestionInput{BankID: uuid.New(), Type: "JENIS_ANEH"})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("MC valid → opsi dibangun + skor dinormalisasi", func(t *testing.T) {
		uc, repo, banks, _, _ := newQuestionUC()
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return bankOf(nil), nil
		}
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Question, error) {
			return &model.Question{}, nil
		}
		var saved *model.Question
		repo.createWithOptionsFn = func(_ context.Context, q *model.Question) error {
			saved = q
			q.ID = uuid.New()
			return nil
		}

		out, err := uc.Create(ctxBg(), QuestionInput{
			BankID: uuid.New(), Type: model.QuestionTypeMultipleChoice,
			Content: "Soal?", ScoreWeight: 0, // → dinormalisasi ke 1.0
			Options: mcOptions(),
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if saved == nil || len(saved.Options) != 2 {
			t.Errorf("opsi tersimpan = %d, want 2", len(saved.Options))
		}
		if saved.ScoreWeight != 1.0 {
			t.Errorf("skor = %v, want 1.0 (normalisasi)", saved.ScoreWeight)
		}
		_ = out
	})

	t.Run("esai tanpa opsi, answer_keys digabung", func(t *testing.T) {
		uc, repo, banks, _, _ := newQuestionUC()
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return bankOf(nil), nil
		}
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Question, error) {
			return &model.Question{}, nil
		}
		var saved *model.Question
		repo.createWithOptionsFn = func(_ context.Context, q *model.Question) error {
			saved = q
			q.ID = uuid.New()
			return nil
		}

		_, err := uc.Create(ctxBg(), QuestionInput{
			BankID: uuid.New(), Type: model.QuestionTypeEssay,
			Content: "Jelaskan!", AnswerKeys: []string{"kunci 1", "kunci 2"},
		})
		if err != nil {
			t.Fatalf("create esai: %v", err)
		}
		if len(saved.Options) != 0 {
			t.Error("esai tidak boleh punya opsi")
		}
		if saved.AnswerKeys == nil || !strings.Contains(*saved.AnswerKeys, "kunci 1") {
			t.Errorf("answer keys = %v", saved.AnswerKeys)
		}
	})
}

func TestQuestionDelete_GuardInUse(t *testing.T) {
	uc, repo, _, sections, answers := newQuestionUC()

	t.Run("soal dipakai section → 409", func(t *testing.T) {
		qid := uuid.New()
		sections.usedQ = map[uuid.UUID]bool{qid: true}
		err := uc.Delete(ctxBg(), qid)
		if ae := apperror.From(err); ae.Code != 409 {
			t.Fatalf("want 409, got %v", err)
		}
	})

	t.Run("soal terjawab siswa → 409", func(t *testing.T) {
		qid := uuid.New()
		sections.usedQ = map[uuid.UUID]bool{}
		answers.answeredOverride = map[uuid.UUID]bool{qid: true}
		err := uc.Delete(ctxBg(), qid)
		if ae := apperror.From(err); ae.Code != 409 {
			t.Fatalf("want 409, got %v", err)
		}
	})

	t.Run("bebas → hapus", func(t *testing.T) {
		qid := uuid.New()
		sections.usedQ = map[uuid.UUID]bool{}
		answers.answeredOverride = map[uuid.UUID]bool{}
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Question, error) {
			return &model.Question{BaseModel: model.BaseModel{ID: qid}}, nil
		}
		if err := uc.Delete(ctxBg(), qid); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if len(repo.deleted) != 1 || repo.deleted[0] != qid {
			t.Error("repo.Delete tidak dipanggil")
		}
	})
}
