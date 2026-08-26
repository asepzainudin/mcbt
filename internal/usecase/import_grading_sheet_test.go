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

// buildSheetXLSX generik: header + baris string.
func buildSheetXLSX(t *testing.T, columns []string, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	for c, col := range columns {
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

func TestTeacherImport(t *testing.T) {
	t.Run("valid → guru dibuat massal", func(t *testing.T) {
		uc, repo, roles := newTeacherUC()
		roles.findByNameFn = func(context.Context, string) (*model.Role, error) {
			return &model.Role{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "teacher"}, nil
		}
		var imported int
		repo.createManyFn = func(_ context.Context, ups []repository.TeacherUpsert) error {
			imported = len(ups)
			return nil
		}
		data := buildSheetXLSX(t, teacherTemplateColumns, [][]string{
			{"guru1", "Guru Satu", "guru1@x.id", "123", "0811"},
			{"guru2", "Guru Dua", "guru2@x.id", "", ""},
		})
		res, err := uc.Import(ctxBg(), data)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if res.ImportedCount != 2 || imported != 2 {
			t.Errorf("imported=%d dibuat=%d", res.ImportedCount, imported)
		}
	})

	t.Run("baris tanpa nama → di-skip", func(t *testing.T) {
		uc, _, roles := newTeacherUC()
		roles.findByNameFn = func(context.Context, string) (*model.Role, error) {
			return &model.Role{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		}
		data := buildSheetXLSX(t, teacherTemplateColumns, [][]string{
			{"guru3", "", "guru3@x.id", "", ""},
		})
		res, err := uc.Import(ctxBg(), data)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if res.ImportedCount != 0 || len(res.Skipped) != 1 {
			t.Errorf("hasil=%+v", res)
		}
	})
}

func TestStudentImport(t *testing.T) {
	t.Run("valid dengan kelas", func(t *testing.T) {
		uc, repo, roles, classes := newStudentUC()
		roles.findByNameFn = func(context.Context, string) (*model.Role, error) {
			return &model.Role{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "student"}, nil
		}
		classes.listAllFn = func(context.Context) ([]model.Class, error) {
			return []model.Class{{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "VII-A"}}, nil
		}
		var imported int
		repo.createManyFn = func(_ context.Context, ups []repository.StudentUpsert) error {
			imported = len(ups)
			return nil
		}
		data := buildSheetXLSX(t, studentTemplateColumns, [][]string{
			{"siswa1", "Siswa Satu", "siswa1@x.id", "20260001", "0812", "VII-A"},
		})
		res, err := uc.Import(ctxBg(), data)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if res.ImportedCount != 1 || imported != 1 {
			t.Errorf("hasil=%+v dibuat=%d", res, imported)
		}
	})
}

func TestExamGradingSheet(t *testing.T) {
	uc, grading, exams := newGradingUC()
	examID := uuid.New()
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{BaseModel: model.BaseModel{ID: examID}}, nil
	}
	attemptID := uuid.New()
	qid := uuid.New()
	grading.submitted = []model.ExamAttempt{{
		BaseModel: model.BaseModel{ID: attemptID}, Status: model.AttemptStatusSubmitted, Score: fltPtr(8),
	}}
	grading.answersByAttempt = map[uuid.UUID][]model.ExamAnswer{
		attemptID: {{AttemptID: attemptID, QuestionID: qid, AnswerValue: "A", Score: fltPtr(8)}},
	}
	grading.questionsByID = map[uuid.UUID]*model.Question{
		qid: {BaseModel: model.BaseModel{ID: qid}, QuestionType: model.QuestionTypeMultipleChoice,
			Content: "Soal sheet?", ScoreWeight: 10, Options: []model.Option{{Label: "A", Content: "opsi A"}}},
	}

	sheet, err := uc.ExamGradingSheet(ctxBg(), examID)
	if err != nil {
		t.Fatalf("sheet: %v", err)
	}
	if len(sheet) != 1 {
		t.Fatalf("siswa = %d, want 1", len(sheet))
	}
	st := sheet[0]
	if st.AttemptID != attemptID || len(st.Answers) != 1 {
		t.Errorf("sheet salah: %+v", st)
	}
	if st.Answers[0].Type != "MULTIPLE_CHOICE" || len(st.Answers[0].OptionTexts) != 1 {
		t.Errorf("detail jawaban salah: %+v", st.Answers[0])
	}
}

// ---------- pelengkap akhir: CRUD kecil & utilitas ----------

func TestQuestionListGetUpdateReorder(t *testing.T) {
	uc, repo, banks, sections, answers := newQuestionUC()
	bankID := uuid.New()
	banks.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
		return bankOf(nil), nil
	}
	qid := uuid.New()
	existing := &model.Question{BaseModel: model.BaseModel{ID: qid}, QuestionBankID: bankID, QuestionType: model.QuestionTypeMultipleChoice}
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Question, error) { return existing, nil }

	if _, _, err := uc.List(ctxBg(), repository.QuestionListParams{}); err != nil {
		t.Errorf("list: %v", err)
	}
	if _, err := uc.Get(ctxBg(), qid); err != nil {
		t.Errorf("get: %v", err)
	}
	sections.usedQ = map[uuid.UUID]bool{}
	answers.answeredOverride = map[uuid.UUID]bool{}
	if _, err := uc.Update(ctxBg(), qid, QuestionInput{Type: model.QuestionTypeMultipleChoice, Content: "baru", Options: mcOptions()}); err != nil {
		t.Errorf("update: %v", err)
	}
	if err := uc.ReorderOptions(ctxBg(), qid, []uuid.UUID{uuid.New()}); err != nil {
		t.Errorf("reorder: %v", err)
	}
	if _, err := uc.SetCorrectOption(ctxBg(), qid, uuid.New()); err != nil {
		t.Errorf("setCorrect: %v", err)
	}
	// short_answer tanpa answer_keys → 422
	_, err := uc.Create(ctxBg(), QuestionInput{BankID: bankID, Type: model.QuestionTypeShortAnswer})
	if apperror.From(err).Code != 422 {
		t.Errorf("short answer wajib kunci: %v", err)
	}
}

func TestTeacherCrud(t *testing.T) {
	uc, repo, roles := newTeacherUC()
	roles.findByNameFn = func(context.Context, string) (*model.Role, error) {
		return &model.Role{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
	}
	if _, total, err := uc.List(ctxBg(), "", 1, 10); err != nil || total != 0 {
		t.Errorf("list: %v (%d)", err, total)
	}
	if _, err := uc.Create(ctxBg(), repository.TeacherUpsert{
		Username: "gurubaru", Name: "Guru Baru", Email: "gb@x.id", PasswordHash: "x",
	}); err != nil {
		t.Errorf("create: %v", err)
	}
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Teacher, error) {
		return teacherWithUser("Guru Edit"), nil
	}
	if _, err := uc.Update(ctxBg(), uuid.New(), repository.TeacherUpdate{Name: "Guru Diedit"}); err != nil {
		t.Errorf("update: %v", err)
	}
	if err := uc.Delete(ctxBg(), uuid.New()); err != nil {
		t.Errorf("delete: %v", err)
	}
	if data, err := uc.TemplateXLSX(); err != nil || len(data) == 0 {
		t.Errorf("template: %v", err)
	}
}

func TestStudentTemplateAndSubjectUpdate(t *testing.T) {
	uc, _, _, _ := newStudentUC()
	if data, err := uc.TemplateXLSX(); err != nil || len(data) == 0 {
		t.Errorf("template: %v", err)
	}
	subjects := &fakeSubjectRepo{}
	subjectUC := NewSubjectUsecase(subjects)
	subjects.findByIDFn = func(context.Context, uuid.UUID) (*model.Subject, error) {
		return &model.Subject{}, nil
	}
	if _, err := subjectUC.Update(ctxBg(), uuid.New(), SubjectInput{Code: "KIM", Name: "Kimia"}); err != nil {
		t.Errorf("subject update: %v", err)
	}
}

func TestExamRefExistsMissing(t *testing.T) {
	uc, _, subjects, ays, _, _ := newExamUC()
	subjects.findByIDFn = func(context.Context, uuid.UUID) (*model.Subject, error) { return nil, notFound() }
	ays.findByIDFn = func(context.Context, uuid.UUID) (*model.AcademicYear, error) { return nil, notFound() }

	// mapel tak ada → 422/404; tahun ajaran tak ada → error juga
	_, err := uc.Create(ctxBg(), ExamInput{Title: "X", SubjectID: uuid.New()})
	if err == nil {
		t.Error("mapel tidak ada harus error")
	}
	subjectOK(subjects)
	ayID := uuid.New()
	_, err = uc.Create(ctxBg(), ExamInput{Title: "X", SubjectID: uuid.New(), AcademicYearID: &ayID})
	if err == nil {
		t.Error("tahun ajaran tidak ada harus error")
	}
}

func TestStudentUpdateFlow(t *testing.T) {
	uc, repo, _, classes := newStudentUC()
	stu := studentWithUser("Edit")
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) { return stu, nil }
	classes.findByIDFn = func(context.Context, uuid.UUID) (*model.Class, error) { return &model.Class{}, nil }

	if _, err := uc.Update(ctxBg(), stu.ID, repository.StudentUpdate{Name: "Nama Edit", Email: "e@x.id", Nis: stu.Nis}); err != nil {
		t.Errorf("update: %v", err)
	}
}

func TestNotFoundBranches(t *testing.T) {
	t.Run("bank get notfound", func(t *testing.T) {
		uc, repo, _, _ := newBankUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) { return nil, notFound() }
		if _, err := uc.Get(ctxBg(), uuid.New()); apperror.From(err).Code != 404 {
			t.Error("want 404")
		}
	})
	t.Run("section update notfound", func(t *testing.T) {
		uc, sections, _, _ := newSectionUC()
		sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) { return nil, notFound() }
		if _, err := uc.Update(ctxBg(), uuid.New(), ExamSectionInput{Name: "X", Sequence: 1}); apperror.From(err).Code != 404 {
			t.Error("want 404")
		}
	})
	t.Run("schedule getByExam notfound", func(t *testing.T) {
		uc, _, exams := newScheduleUC()
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		if _, err := uc.GetByExam(ctxBg(), uuid.New()); apperror.From(err).Code != 404 {
			t.Error("want 404")
		}
	})
	t.Run("candidate exam notfound", func(t *testing.T) {
		f := newCandidateFixture()
		f.exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		if err := f.uc.ValidateToken(ctxBg(), f.userID, uuid.New(), ""); apperror.From(err).Code != 404 {
			t.Error("want 404")
		}
	})
	t.Run("grading sheet notfound", func(t *testing.T) {
		uc, _, exams := newGradingUC()
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		if _, err := uc.ExamGradingSheet(ctxBg(), uuid.New()); apperror.From(err).Code != 404 {
			t.Error("want 404")
		}
	})
	t.Run("student update notfound", func(t *testing.T) {
		uc, repo, _, _ := newStudentUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) { return nil, notFound() }
		if _, err := uc.Update(ctxBg(), uuid.New(), repository.StudentUpdate{Name: "x", Email: "e@x.id", Nis: "1"}); apperror.From(err).Code != 404 {
			t.Error("want 404")
		}
	})
}
