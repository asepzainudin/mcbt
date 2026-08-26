package usecase

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

// buildImportXLSX membuat file Excel impor soal dari baris teks 10 kolom.
func buildImportXLSX(t *testing.T, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	for c, col := range questionImportColumns {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		_ = f.SetCellValue("Sheet1", cell, col)
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue("Sheet1", cell, v)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("build xlsx: %v", err)
	}
	return buf.Bytes()
}

func TestQuestionImportTemplate(t *testing.T) {
	uc := NewQuestionImportUsecase(NewImportTokenStore(), &fakeQuestionRepo{})
	data, err := uc.TemplateXLSX()
	if err != nil {
		t.Fatalf("template: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("template kosong")
	}
	// template harus lolos parseSheet (baris sample → 5 data)
	rows, err := parseSheet(data, questionImportColumns)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	if len(rows) != 5 {
		t.Errorf("baris sample = %d, want 5", len(rows))
	}
}

func TestQuestionImportValidate(t *testing.T) {
	uc := NewQuestionImportUsecase(NewImportTokenStore(), &fakeQuestionRepo{})
	bankID := uuid.New()

	t.Run("valid → token + jumlah baris", func(t *testing.T) {
		data := buildImportXLSX(t, [][]string{
			{"multiple_choice", "Soal PG?", "2", "pembahasan", "A", "B", "", "", "", "A"},
			{"essay", "Soal esai?", "5", "", "", "", "", "", "", "kunci"},
		})
		token, count, skipped, err := uc.Validate(ctxBg(), data, bankID)
		if err != nil {
			t.Fatalf("validate: %v", err)
		}
		if count != 2 || token == "" || len(skipped) != 0 {
			t.Errorf("count=%d token=%q skipped=%+v", count, token, skipped)
		}
	})

	t.Run("tipe tidak valid → baris di-skip", func(t *testing.T) {
		data := buildImportXLSX(t, [][]string{
			{"jenis_aneh", "Soal?", "1", "", "", "", "", "", "", "A"},
		})
		_, count, skipped, err := uc.Validate(ctxBg(), data, bankID)
		if apperror.From(err).Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
		if count != 0 || len(skipped) != 1 || skipped[0].Field != "type" {
			t.Errorf("count=%d skipped=%+v", count, skipped)
		}
	})

	t.Run("file kosong → 422", func(t *testing.T) {
		data := buildImportXLSX(t, nil)
		_, _, _, err := uc.Validate(ctxBg(), data, bankID)
		if apperror.From(err).Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})
}

func TestQuestionImportProcess(t *testing.T) {
	t.Run("token valid → soal diimpor", func(t *testing.T) {
		store := NewImportTokenStore()
		repo := &fakeQuestionRepo{}
		uc := NewQuestionImportUsecase(store, repo)
		bankID := uuid.New()

		data := buildImportXLSX(t, [][]string{
			{"multiple_choice", "Soal?", "1", "", "A", "B", "", "", "", "A"},
		})
		token, _, _, err := uc.Validate(ctxBg(), data, bankID)
		if err != nil {
			t.Fatalf("validate: %v", err)
		}

		res, err := uc.Process(ctxBg(), token)
		if err != nil {
			t.Fatalf("process: %v", err)
		}
		if res.ImportedCount != 1 {
			t.Errorf("imported = %d, want 1", res.ImportedCount)
		}
		if len(repo.imported) != 1 || repo.imported[0].QuestionBankID != bankID {
			t.Errorf("impor salah: %+v", repo.imported)
		}
	})

	t.Run("token kedaluwarsa → 422", func(t *testing.T) {
		uc := NewQuestionImportUsecase(NewImportTokenStore(), &fakeQuestionRepo{})
		_, err := uc.Process(ctxBg(), "token-palsu")
		if apperror.From(err).Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("token hanya bisa dipakai sekali", func(t *testing.T) {
		store := NewImportTokenStore()
		repo := &fakeQuestionRepo{}
		uc := NewQuestionImportUsecase(store, repo)
		data := buildImportXLSX(t, [][]string{
			{"true_false", "Bumi datar.", "1", "", "BENAR", "SALAH", "", "", "", "B"},
		})
		token, _, _, _ := uc.Validate(ctxBg(), data, uuid.New())
		if _, err := uc.Process(ctxBg(), token); err != nil {
			t.Fatalf("proses pertama: %v", err)
		}
		if _, err := uc.Process(ctxBg(), token); err == nil {
			t.Error("token kedua kalinya harus gagal")
		}
	})
}

// ---------- pelengkap coverage kecil ----------

func TestProfileGet(t *testing.T) {
	profiles := &fakeProfileRepo{}
	users := &fakeUserRepo{}
	uc := NewProfileUsecase(profiles, users)

	out, err := uc.Get(ctxBg(), uuid.New())
	if err != nil || out == nil {
		t.Errorf("get: %v", err)
	}
}

func TestBankGetAndUpdate(t *testing.T) {
	uc, repo, subjects, _ := newBankUC()
	bankID := uuid.New()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
		return bankOf(nil), nil
	}
	if _, err := uc.Get(ctxBg(), bankID); err != nil {
		t.Errorf("get: %v", err)
	}
	subjectOK(subjects)
	if _, err := uc.Update(ctxBg(), bankID, QuestionBankInput{SubjectID: uuid.New(), Code: "X", Title: "Baru"}); err != nil {
		t.Errorf("update: %v", err)
	}
}

func TestClassListUpdateDelete(t *testing.T) {
	classRepo := &fakeClassRepo{}
	ayRepo := &fakeAYRepo{}
	uc := NewClassUsecase(classRepo, ayRepo)
	ayID := uuid.New()
	ayRepo.findByIDFn = func(context.Context, uuid.UUID) (*model.AcademicYear, error) {
		return &model.AcademicYear{BaseModel: model.BaseModel{ID: ayID}}, nil
	}

	if _, _, err := uc.List(ctxBg(), "", nil, 1, 10); err != nil {
		t.Errorf("list: %v", err)
	}
	if _, err := uc.Create(ctxBg(), ClassInput{Name: "VIII-B", AcademicYearID: ayID}); err != nil {
		t.Errorf("create: %v", err)
	}
	classRepo.findByIDFn = func(context.Context, uuid.UUID) (*model.Class, error) {
		return &model.Class{BaseModel: model.BaseModel{ID: uuid.New()}, AcademicYearID: ayID}, nil
	}
	if _, err := uc.Update(ctxBg(), uuid.New(), ClassInput{Name: "VIII-C", AcademicYearID: ayID}); err != nil {
		t.Errorf("update: %v", err)
	}
	if err := uc.Delete(ctxBg(), uuid.New()); err != nil {
		t.Errorf("delete: %v", err)
	}
}

func TestAYAndSubjectList(t *testing.T) {
	ayUC := NewAcademicYearUsecase(&fakeAYRepo{})
	if _, _, err := ayUC.List(ctxBg(), "", 1, 10); err != nil {
		t.Errorf("ay list: %v", err)
	}
	subjectUC := NewSubjectUsecase(&fakeSubjectRepo{})
	if _, _, err := subjectUC.List(ctxBg(), "", 1, 10); err != nil {
		t.Errorf("subject list: %v", err)
	}
}

func TestSectionUpdateAndExamQuestionsGroups(t *testing.T) {
	uc, sections, exams, _ := newSectionUC()
	examID := uuid.New()
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{BaseModel: model.BaseModel{ID: examID}}, nil
	}
	sectionID := uuid.New()
	sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) {
		return &model.ExamSection{BaseModel: model.BaseModel{ID: sectionID}, ExamID: examID}, nil
	}

	if _, err := uc.Update(ctxBg(), sectionID, ExamSectionInput{Name: "Esai", Sequence: 2}); err != nil {
		t.Errorf("update: %v", err)
	}
	sections.listExamQFn = func(context.Context, *model.Exam) ([]repository.ExamQuestionGroup, error) {
		return []repository.ExamQuestionGroup{{
			Section:   model.ExamSection{Name: "PG", Sequence: 1},
			Questions: []model.Question{{BaseModel: model.BaseModel{ID: uuid.New()}, QuestionType: model.QuestionTypeMultipleChoice}},
		}}, nil
	}
	groups, err := uc.ExamQuestions(ctxBg(), examID)
	if err != nil || len(groups) != 1 {
		t.Errorf("examQuestions: %v (%d)", err, len(groups))
	}
}
