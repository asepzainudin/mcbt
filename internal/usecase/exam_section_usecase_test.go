package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

func newExamUC() (*ExamUsecase, *fakeExamRepo, *fakeSubjectRepo, *fakeAYRepo, *fakeBankRepo, *fakeAttemptRepo) {
	repo := &fakeExamRepo{}
	subjects := &fakeSubjectRepo{}
	ays := &fakeAYRepo{}
	banks := &fakeBankRepo{}
	attempts := &fakeAttemptRepo{}
	uc := NewExamUsecase(repo, subjects, ays, banks, attempts)
	return uc, repo, subjects, ays, banks, attempts
}

func validSettings() ExamSettingsInput {
	return ExamSettingsInput{
		DurationMinutes: 60, MaxAttempts: 1, PassingGrade: 70,
		AllowBacktrack: true, AutoSubmit: true,
	}
}

func subjectOK(subjects *fakeSubjectRepo) {
	subjects.findByIDFn = func(context.Context, uuid.UUID) (*model.Subject, error) {
		return &model.Subject{}, nil
	}
}

func TestExamCreate_SetsDraftAndCreator(t *testing.T) {
	uc, repo, subjects, _, _, _ := newExamUC()
	subjectOK(subjects)
	creator := uuid.New()
	repo.findByIDFn = func(_ context.Context, _ uuid.UUID) (*model.Exam, error) {
		return repo.created[len(repo.created)-1], nil
	}

	out, err := uc.Create(ctxBg(), ExamInput{Title: "Ujian 1", SubjectID: uuid.New(), CreatedBy: &creator})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatal("repo.Create tidak dipanggil")
	}
	saved := repo.created[0]
	if saved.Status != model.ExamStatusDraft {
		t.Errorf("status = %q, want draft", saved.Status)
	}
	if saved.CreatedBy == nil || *saved.CreatedBy != creator {
		t.Error("CreatedBy tidak di-set")
	}
	if out != saved {
		t.Error("harus kembalikan exam yang sama via FindByID")
	}
}

func TestExamList_AttachesAttemptsCount(t *testing.T) {
	uc, repo, _, _, _, attempts := newExamUC()
	e1 := model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}}
	e2 := model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}}
	repo.listFn = func(context.Context, repository.ExamListParams) (*repository.PageResult[model.Exam], error) {
		return pageOf([]model.Exam{e1, e2}), nil
	}
	attempts.counts = map[string]int64{
		e1.ID.String(): 7,
		e2.ID.String(): 0,
	}

	items, total, err := uc.List(ctxBg(), repository.ExamListParams{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("items=%d total=%d", len(items), total)
	}
	if items[0].AttemptsCount != 7 {
		t.Errorf("attempts_count[0] = %d, want 7", items[0].AttemptsCount)
	}
}

func TestExamPublishClose(t *testing.T) {
	uc, repo, _, _, _, _ := newExamUC()
	id := uuid.New()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{BaseModel: model.BaseModel{ID: id}}, nil
	}

	if _, err := uc.Publish(ctxBg(), id); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if repo.setStatuses[0].status != model.ExamStatusPublished {
		t.Errorf("status = %q", repo.setStatuses[0].status)
	}
	if _, err := uc.Close(ctxBg(), id); err != nil {
		t.Fatalf("close: %v", err)
	}
	if repo.setStatuses[1].status != model.ExamStatusClosed {
		t.Errorf("status = %q", repo.setStatuses[1].status)
	}
}

func TestExamDelete_GuardUsedByParticipants(t *testing.T) {
	uc, repo, _, _, _, attempts := newExamUC()
	id := uuid.New()
	attempts.counts = map[string]int64{id.String(): 3}

	err := uc.Delete(ctxBg(), id)
	if ae := apperror.From(err); ae.Code != 409 {
		t.Fatalf("ujian terpakai want 409, got %v", err)
	}
	if len(repo.deleted) != 0 {
		t.Error("tidak boleh menghapus")
	}

	attempts.counts = map[string]int64{}
	if err := uc.Delete(ctxBg(), id); err != nil {
		t.Fatalf("hapus bebas: %v", err)
	}
	if len(repo.deleted) != 1 {
		t.Error("repo.Delete tidak dipanggil")
	}
}

func TestExamUpdateSettings_TokenBehavior(t *testing.T) {
	uc, repo, _, _, _, _ := newExamUC()
	examID := uuid.New()
	exam := &model.Exam{BaseModel: model.BaseModel{ID: examID}}
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return exam, nil }

	t.Run("tokenEnabled → token digenerate", func(t *testing.T) {
		exam.ExamToken = nil
		in := validSettings()
		in.TokenEnabled = true
		if _, err := uc.UpdateSettings(ctxBg(), examID, in); err != nil {
			t.Fatalf("settings: %v", err)
		}
		if exam.ExamToken == nil || len(*exam.ExamToken) < 6 {
			t.Errorf("token = %v, want >= 6 karakter", exam.ExamToken)
		}
	})

	t.Run("tokenOff → token dibuang", func(t *testing.T) {
		tok := "ABC123"
		exam.ExamToken = &tok
		in := validSettings()
		in.TokenEnabled = false
		if _, err := uc.UpdateSettings(ctxBg(), examID, in); err != nil {
			t.Fatalf("settings: %v", err)
		}
		if exam.ExamToken != nil {
			t.Error("token harus dibuang saat nonaktif")
		}
	})

	t.Run("ujian tak ada → 404", func(t *testing.T) {
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		_, err := uc.UpdateSettings(ctxBg(), uuid.New(), ExamSettingsInput{})
		if ae := apperror.From(err).Code; ae != 404 {
			t.Fatalf("want 404, got %v", ae)
		}
	})
}

// ---------- ExamSectionUsecase ----------

func newSectionUC() (*ExamSectionUsecase, *fakeSectionRepo, *fakeExamRepo, *fakeBankRepo) {
	sections := &fakeSectionRepo{}
	exams := &fakeExamRepo{}
	banks := &fakeBankRepo{}
	return NewExamSectionUsecase(sections, exams, banks), sections, exams, banks
}

func TestSectionCreate_Validation(t *testing.T) {
	uc, _, exams, _ := newSectionUC()
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{}, nil
	}

	t.Run("nama kosong → 422", func(t *testing.T) {
		_, err := uc.Create(ctxBg(), uuid.New(), ExamSectionInput{Name: "  ", Sequence: 1})
		ae := apperror.From(err)
		if ae.Code != 422 || ae.Details["name"] == "" {
			t.Fatalf("want 422 name, got %v", err)
		}
	})

	t.Run("sequence 0 → 422", func(t *testing.T) {
		_, err := uc.Create(ctxBg(), uuid.New(), ExamSectionInput{Name: "PG", Sequence: 0})
		ae := apperror.From(err)
		if ae.Code != 422 || ae.Details["sequence"] == "" {
			t.Fatalf("want 422 sequence, got %v", err)
		}
	})

	t.Run("valid → tersimpan", func(t *testing.T) {
		out, err := uc.Create(ctxBg(), uuid.New(), ExamSectionInput{Name: "PG", Sequence: 1})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if out.Name != "PG" || out.Sequence != 1 {
			t.Errorf("hasil tidak sesuai: %+v", out)
		}
	})
}

func TestMapQuestions(t *testing.T) {
	t.Run("tanpa bank → 422", func(t *testing.T) {
		uc, _, _, _ := newSectionUC()
		_, _, err := uc.MapQuestions(ctxBg(), uuid.New(), MapQuestionsInput{})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("bank tak ada → 422 detail", func(t *testing.T) {
		uc, sections, _, banks := newSectionUC()
		sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) {
			return &model.ExamSection{}, nil
		}
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return nil, notFound()
		}
		_, _, err := uc.MapQuestions(ctxBg(), uuid.New(), MapQuestionsInput{BankIDs: []uuid.UUID{uuid.New()}})
		ae := apperror.From(err)
		if ae.Code != 422 || ae.Details["question_bank_ids"] != "bank soal tidak ditemukan" {
			t.Fatalf("want 422 bank, got %v", err)
		}
	})

	t.Run("bank kosong soal → 422", func(t *testing.T) {
		uc, sections, _, banks := newSectionUC()
		sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) {
			return &model.ExamSection{}, nil
		}
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return &model.QuestionBank{}, nil
		}
		sections.qidsByBanks = []uuid.UUID{} // tak ada soal
		_, _, err := uc.MapQuestions(ctxBg(), uuid.New(), MapQuestionsInput{BankIDs: []uuid.UUID{uuid.New()}})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("mapping: kandidat disaring & skipped dihitung", func(t *testing.T) {
		uc, sections, _, banks := newSectionUC()
		sectionID := uuid.New()
		sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) {
			return &model.ExamSection{BaseModel: model.BaseModel{ID: sectionID}, ExamID: uuid.New()}, nil
		}
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return &model.QuestionBank{}, nil
		}
		q1, q2, q3 := uuid.New(), uuid.New(), uuid.New()
		sections.qidsByBanks = []uuid.UUID{q1, q2, q3}    // tersedia 3
		sections.mappedSet = map[uuid.UUID]bool{q1: true} // q1 sudah termapping

		inserted, skipped, err := uc.MapQuestions(ctxBg(), sectionID, MapQuestionsInput{
			BankIDs: []uuid.UUID{uuid.New()},
		})
		if err != nil {
			t.Fatalf("map: %v", err)
		}
		if inserted != 2 {
			t.Errorf("inserted = %d, want 2 (q2,q3)", inserted)
		}
		if skipped != 1 {
			t.Errorf("skipped = %d, want 1 (q1)", skipped)
		}
		if len(sections.inserted) != 1 || len(sections.inserted[0].qids) != 2 {
			t.Errorf("InsertMappings salah: %+v", sections.inserted)
		}
	})

	t.Run("total_random membatasi kandidat", func(t *testing.T) {
		uc, sections, _, banks := newSectionUC()
		sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) {
			return &model.ExamSection{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		}
		banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return &model.QuestionBank{}, nil
		}
		sections.qidsByBanks = []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
		sections.mappedSet = map[uuid.UUID]bool{}

		inserted, _, err := uc.MapQuestions(ctxBg(), uuid.New(), MapQuestionsInput{
			BankIDs: []uuid.UUID{uuid.New()}, TotalRandomQuestions: 2,
		})
		if err != nil {
			t.Fatalf("map: %v", err)
		}
		if inserted != 2 {
			t.Errorf("inserted = %d, want 2", inserted)
		}
	})
}
