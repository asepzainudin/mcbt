package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

// ---------- GradingUsecase ----------

func newGradingUC() (*GradingUsecase, *fakeGradingRepo, *fakeExamRepo) {
	grading := &fakeGradingRepo{answersByAttempt: map[uuid.UUID][]model.ExamAnswer{}}
	exams := &fakeExamRepo{}
	uc := NewGradingUsecase(grading, &fakeAnswerRepo{}, attemptsFake(), exams)
	return uc, grading, exams
}

func attemptsFake() *fakeAttemptRepo { return &fakeAttemptRepo{} }

func TestCalculateGrades(t *testing.T) {
	t.Run("ujian tak ada → 404", func(t *testing.T) {
		uc, _, exams := newGradingUC()
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		_, err := uc.CalculateGrades(ctxBg(), uuid.New())
		if ae := apperror.From(err).Code; ae != 404 {
			t.Fatalf("want 404, got %v", ae)
		}
	})

	t.Run("objektif dinilai, esai dilewati, total dari DB", func(t *testing.T) {
		uc, grading, exams := newGradingUC()
		examID := uuid.New()
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
			return &model.Exam{BaseModel: model.BaseModel{ID: examID}}, nil
		}
		attemptID := uuid.New()
		grading.submitted = []model.ExamAttempt{{BaseModel: model.BaseModel{ID: attemptID}}}

		qMC := uuid.New()
		qEssay := uuid.New()
		grading.questionsByID = map[uuid.UUID]*model.Question{
			qMC:    {BaseModel: model.BaseModel{ID: qMC}, QuestionType: model.QuestionTypeMultipleChoice, ScoreWeight: 8, Options: []model.Option{{Label: "A", IsCorrect: true}}},
			qEssay: {BaseModel: model.BaseModel{ID: qEssay}, QuestionType: model.QuestionTypeEssay, ScoreWeight: 20},
		}
		grading.answersByAttempt[attemptID] = []model.ExamAnswer{
			{BaseModel: model.BaseModel{ID: uuid.New()}, QuestionID: qMC, AnswerValue: "A"},
			{BaseModel: model.BaseModel{ID: uuid.New()}, QuestionID: qEssay, AnswerValue: "esai siswa"},
		}
		grading.sumScores = map[uuid.UUID]float64{attemptID: 8} // total hasil rekap DB

		res, err := uc.CalculateGrades(ctxBg(), examID)
		if err != nil {
			t.Fatalf("calculate: %v", err)
		}
		if res.AttemptsGraded != 1 || res.QuestionsGraded != 1 {
			t.Errorf("hasil = %+v, want 1 attempt & 1 soal objektif", res)
		}
		if len(grading.updatedScore) != 1 || grading.updatedScore[0].score != 8 {
			t.Errorf("skor total = %+v, want 8", grading.updatedScore)
		}
	})
}

func TestGradeEssay(t *testing.T) {
	t.Run("bukan esai → 422", func(t *testing.T) {
		uc, grading, _ := newGradingUC()
		attemptID, qid := uuid.New(), uuid.New()
		grading.findAnswerByAttemptFn = func(context.Context, uuid.UUID, uuid.UUID) (*model.ExamAnswer, error) {
			return &model.ExamAnswer{}, nil
		}
		grading.questionsByID = map[uuid.UUID]*model.Question{
			qid: {BaseModel: model.BaseModel{ID: qid}, QuestionType: model.QuestionTypeMultipleChoice},
		}
		_, err := uc.GradeEssay(ctxBg(), attemptID, qid, GradeEssayInput{Score: 5})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("skor melebihi bobot → 422", func(t *testing.T) {
		uc, grading, _ := newGradingUC()
		attemptID, qid := uuid.New(), uuid.New()
		grading.findAnswerByAttemptFn = func(context.Context, uuid.UUID, uuid.UUID) (*model.ExamAnswer, error) {
			return &model.ExamAnswer{}, nil
		}
		grading.questionsByID = map[uuid.UUID]*model.Question{
			qid: {BaseModel: model.BaseModel{ID: qid}, QuestionType: model.QuestionTypeEssay, ScoreWeight: 10},
		}
		_, err := uc.GradeEssay(ctxBg(), attemptID, qid, GradeEssayInput{Score: 11})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("valid → update + hitung ulang total", func(t *testing.T) {
		uc, grading, _ := newGradingUC()
		attemptID, qid, answerID := uuid.New(), uuid.New(), uuid.New()
		grading.findAnswerByAttemptFn = func(context.Context, uuid.UUID, uuid.UUID) (*model.ExamAnswer, error) {
			return &model.ExamAnswer{BaseModel: model.BaseModel{ID: answerID}}, nil
		}
		grading.questionsByID = map[uuid.UUID]*model.Question{
			qid: {BaseModel: model.BaseModel{ID: qid}, QuestionType: model.QuestionTypeEssay, ScoreWeight: 20},
		}
		grading.findAnswerByIDFn = func(_ context.Context, id uuid.UUID) (*model.ExamAnswer, error) {
			return &model.ExamAnswer{BaseModel: model.BaseModel{ID: id}}, nil
		}
		grading.sumScores = map[uuid.UUID]float64{attemptID: 15}

		out, err := uc.GradeEssay(ctxBg(), attemptID, qid, GradeEssayInput{Score: 15, Feedback: strPtr2("bagus")})
		if err != nil {
			t.Fatalf("grade: %v", err)
		}
		if len(grading.updatedGrading) != 1 || grading.updatedGrading[0] != answerID {
			t.Error("UpdateGrading tidak tepat")
		}
		if len(grading.updatedScore) != 1 || grading.updatedScore[0].score != 15 {
			t.Errorf("total = %+v, want 15", grading.updatedScore)
		}
		if out == nil {
			t.Error("jawaban terupdate harus dikembalikan")
		}
	})
}

func strPtr2(s string) *string { return &s }

// ---------- ResultUsecase ----------

func newResultUC() (*ResultUsecase, *fakeResultRepo, *fakeExamRepo, *fakeStudentRepo) {
	results := &fakeResultRepo{}
	exams := &fakeExamRepo{}
	students := &fakeStudentRepo{}
	return NewResultUsecase(results, exams, students), results, exams, students
}

func TestResultExamResults_Ranking(t *testing.T) {
	uc, results, exams, _ := newResultUC()
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{}, nil
	}
	score := 80.0
	results.examResults = []repository.ExamResultRow{
		{StudentName: "A", Score: &score, PassingGrade: 70},
		{StudentName: "B", Score: nil, PassingGrade: 70},
	}

	ranked, err := uc.ExamResults(ctxBg(), uuid.New(), nil)
	if err != nil {
		t.Fatalf("results: %v", err)
	}
	if ranked[0].Rank != 1 || ranked[1].Rank != 2 {
		t.Errorf("rank = %d,%d", ranked[0].Rank, ranked[1].Rank)
	}
	if !ranked[0].Passed {
		t.Error("80 >= 70 harus lulus")
	}
	if ranked[1].Passed {
		t.Error("skor nil tidak lulus")
	}
}

func TestResultStudentResults_HideUnpublished(t *testing.T) {
	uc, results, exams, students := newResultUC()
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{}, nil
	}
	students.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) {
		return &model.Student{}, nil
	}
	hidden := repository.StudentResultRow{ExamTitle: "A", ResultsPublished: false, ShowResultImmediately: false, Score: fltPtr(90)}
	shown := repository.StudentResultRow{ExamTitle: "B", ResultsPublished: true, Score: fltPtr(80)}
	results.studentResults = []repository.StudentResultRow{hidden, shown}

	rows, err := uc.StudentResults(ctxBg(), uuid.New())
	if err != nil {
		t.Fatalf("student results: %v", err)
	}
	if rows[0].Score != nil {
		t.Error("skor belum dipublikasikan harus disembunyikan")
	}
	if rows[1].Score == nil || *rows[1].Score != 80 {
		t.Errorf("skor terpublikasi harus tampil: %v", rows[1].Score)
	}
}

func TestResultPublish(t *testing.T) {
	uc, results, exams, _ := newResultUC()
	exams.findByIDFn = func(_ context.Context, _ uuid.UUID) (*model.Exam, error) { return &model.Exam{}, nil }

	if err := uc.PublishResults(ctxBg(), uuid.New(), true); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if len(results.published) != 1 || !results.published[0].val {
		t.Error("SetResultsPublished(true) tidak dipanggil")
	}
}

// ---------- QuestionReportUsecase ----------

func newReportUC() (*QuestionReportUsecase, *fakeReportRepo, *fakeAttemptRepo, *fakeStudentRepo, *fakeOwnerAssertor) {
	reports := &fakeReportRepo{}
	attempts := &fakeAttemptRepo{}
	students := &fakeStudentRepo{}
	access := &fakeOwnerAssertor{}
	uc := NewQuestionReportUsecase(reports, attempts, students, access)
	return uc, reports, attempts, students, access
}

func TestReportCreate(t *testing.T) {
	t.Run("alasan kosong → 422", func(t *testing.T) {
		uc, _, _, _, _ := newReportUC()
		_, err := uc.Create(ctxBg(), CreateReportInput{UserID: uuid.New(), Reason: "   "})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("bukan siswa → 403", func(t *testing.T) {
		uc, _, _, students, _ := newReportUC()
		students.findByUID = func(context.Context, uuid.UUID) (*model.Student, error) {
			return nil, notFound()
		}
		_, err := uc.Create(ctxBg(), CreateReportInput{UserID: uuid.New(), Reason: "kunci salah"})
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})

	t.Run("duplikat → kembalikan existing", func(t *testing.T) {
		uc, reports, _, students, _ := newReportUC()
		students.findByUID = func(context.Context, uuid.UUID) (*model.Student, error) {
			return &model.Student{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		}
		existing := &model.QuestionReport{BaseModel: model.BaseModel{ID: uuid.New()}}
		reports.byAttemptQ = existing

		out, err := uc.Create(ctxBg(), CreateReportInput{AttemptID: uuid.New(), QuestionID: uuid.New(), UserID: uuid.New(), Reason: "dup"})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if out.ID != existing.ID {
			t.Error("harus kembalikan laporan existing")
		}
		if len(reports.created) != 0 {
			t.Error("tidak boleh buat baru")
		}
	})

	t.Run("sukses → status pending", func(t *testing.T) {
		uc, reports, _, students, _ := newReportUC()
		_ = reports
		students.findByUID = func(context.Context, uuid.UUID) (*model.Student, error) {
			return &model.Student{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		}

		out, err := uc.Create(ctxBg(), CreateReportInput{AttemptID: uuid.New(), QuestionID: uuid.New(), UserID: uuid.New(), Reason: "kunci salah"})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if out.Status != model.ReportStatusPending {
			t.Errorf("status = %q, want pending", out.Status)
		}
	})
}

func TestReportList_ScopeOwner(t *testing.T) {
	uc, reports, _, _, _ := newReportUC()
	owner := uuid.New()
	reports.rowsOwner = map[uuid.UUID][]repository.ReportRow{
		owner: {{ID: uuid.New(), ExamTitle: "Milik Guru"}},
	}

	rows, err := uc.List(ctxBg(), "pending", &owner)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if reports.lastOwnerFl == nil || *reports.lastOwnerFl != owner {
		t.Error("filter owner tidak diteruskan ke repo")
	}
}

func TestReportResolve(t *testing.T) {
	reportID := uuid.New()
	resolver := uuid.New()

	t.Run("status tidak valid → 422", func(t *testing.T) {
		uc, reports, _, _, _ := newReportUC()
		reports.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionReport, error) {
			return &model.QuestionReport{}, nil
		}
		_, err := uc.Resolve(ctxBg(), reportID, resolver, ResolveInput{Status: "aneh"}, true)
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("laporan tak ada → 404", func(t *testing.T) {
		uc, reports, _, _, _ := newReportUC()
		reports.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionReport, error) {
			return nil, notFound()
		}
		_, err := uc.Resolve(ctxBg(), reportID, resolver, ResolveInput{Status: "resolved"}, true)
		if ae := apperror.From(err); ae.Code != 404 {
			t.Fatalf("want 404, got %v", err)
		}
	})

	t.Run("bukan pemilik ujian → 403", func(t *testing.T) {
		uc, reports, _, _, access := newReportUC()
		reports.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionReport, error) {
			return &model.QuestionReport{}, nil
		}
		access.err = apperror.Forbidden("bukan milik", nil)
		_, err := uc.Resolve(ctxBg(), reportID, resolver, ResolveInput{Status: "resolved"}, false)
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
		if access.lastAdmin {
			t.Error("isAdmin harus false untuk guru")
		}
	})

	t.Run("sukses → status+resolusi+resolver tersimpan", func(t *testing.T) {
		uc, reports, _, _, _ := newReportUC()
		reports.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionReport, error) {
			return &model.QuestionReport{}, nil
		}
		out, err := uc.Resolve(ctxBg(), reportID, resolver, ResolveInput{
			Status: "resolved", Resolution: strPtr2("sudah dicek"),
		}, true)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if out.Status != "resolved" || out.Resolution == nil || *out.Resolution != "sudah dicek" {
			t.Errorf("hasil resolve salah: %+v", out)
		}
		if out.ResolvedBy == nil || *out.ResolvedBy != resolver {
			t.Error("ResolvedBy tidak di-set")
		}
		if out.ResolvedAt == nil {
			t.Error("ResolvedAt tidak di-set")
		}
	})
}

var _ = gorm.ErrRecordNotFound
