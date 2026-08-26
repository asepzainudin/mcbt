package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

// ---------- harness ----------

type attemptFixture struct {
	uc        *AttemptEngineUsecase
	base      time.Time // acuan ExpiresAt (tidak bergeser)
	userID    uuid.UUID
	studentID uuid.UUID
	examID    uuid.UUID
	attemptID uuid.UUID
	attempts  *fakeAttemptRepo
	answers   *fakeAnswerRepo
	sections  *fakeSectionRepo
	exams     *fakeExamRepo
	grading   *fakeGradingRepo
	now       time.Time
}

func newAttemptFixture(fixed time.Time) *attemptFixture {
	f := &attemptFixture{
		userID: uuid.New(), studentID: uuid.New(), examID: uuid.New(),
		attemptID: uuid.New(), now: fixed, base: fixed,
	}
	attempts := &fakeAttemptRepo{}
	answers := &fakeAnswerRepo{byAttempt: map[uuid.UUID][]model.ExamAnswer{}}
	students := &fakeStudentRepo{}
	sections := &fakeSectionRepo{}
	exams := &fakeExamRepo{}
	grading := &fakeGradingRepo{}

	attempts.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
		if id == f.attemptID {
			return f.currentAttempt(), nil
		}
		return nil, notFound()
	}
	students.findByUID = func(_ context.Context, uid uuid.UUID) (*model.Student, error) {
		if uid == f.userID {
			return &model.Student{BaseModel: model.BaseModel{ID: f.studentID}}, nil
		}
		return nil, notFound()
	}
	exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
		if id == f.examID {
			return &model.Exam{BaseModel: model.BaseModel{ID: f.examID}}, nil
		}
		return nil, notFound()
	}
	sections.inExam = map[string]bool{} // default: tidak ada soal

	f.uc = NewAttemptEngineUsecase(attempts, answers, students, sections, exams, grading)
	f.uc.now = func() time.Time { return f.now }
	f.attempts, f.answers, f.sections, f.exams, f.grading = attempts, answers, sections, exams, grading
	return f
}

// gormNotFoundErr dipisah agar tidak import gorm berulang-ulang di file ini
func (f *attemptFixture) currentAttempt() *model.ExamAttempt {
	return &model.ExamAttempt{
		BaseModel: model.BaseModel{ID: f.attemptID},
		ExamID:    f.examID, StudentID: f.studentID,
		Status:    model.AttemptStatusInProgress,
		StartedAt: f.base.Add(-30 * time.Minute),
		ExpiresAt: f.base.Add(30 * time.Minute),
	}
}

func TestAttemptResolve_Ownership(t *testing.T) {
	f := newAttemptFixture(time.Now())

	t.Run("attempt orang lain → 403", func(t *testing.T) {
		_, err := f.uc.SaveAnswer(ctxBg(), uuid.New() /* user beda */, f.attemptID, SaveAnswerInput{})
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})

	t.Run("attempt tak ada → 404", func(t *testing.T) {
		_, err := f.uc.SaveAnswer(ctxBg(), f.userID, uuid.New(), SaveAnswerInput{})
		if ae := apperror.From(err); ae.Code != 404 {
			t.Fatalf("want 404, got %v", err)
		}
	})
}

func TestAttemptSaveAnswer(t *testing.T) {
	t.Run("valid → upsert dipanggil", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		qid := uuid.New()
		f.sections.inExam[qid.String()] = true

		if _, err := f.uc.SaveAnswer(ctxBg(), f.userID, f.attemptID, SaveAnswerInput{QuestionID: qid, AnswerValue: "A"}); err != nil {
			t.Fatalf("save: %v", err)
		}
		if len(f.answers.upserted) != 1 || f.answers.upserted[0].value != "A" {
			t.Errorf("upsert salah: %+v", f.answers.upserted)
		}
	})

	t.Run("soal di luar ujian → 422", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		_, err := f.uc.SaveAnswer(ctxBg(), f.userID, f.attemptID, SaveAnswerInput{QuestionID: uuid.New()})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("waktu habis → 403 + ditandai expired", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		f.now = f.now.Add(2 * time.Hour) // lewat ExpiresAt
		qid := uuid.New()
		f.sections.inExam[qid.String()] = true

		_, err := f.uc.SaveAnswer(ctxBg(), f.userID, f.attemptID, SaveAnswerInput{QuestionID: qid})
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
		if len(f.attempts.marked) != 1 {
			t.Error("MarkExpired tidak dipanggil")
		}
	})
}

func TestAttemptHeartbeat(t *testing.T) {
	base := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	f := newAttemptFixture(base)

	_, hb, err := f.uc.Heartbeat(ctxBg(), f.userID, f.attemptID)
	if err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if hb.RemainingSeconds != 30*60 {
		t.Errorf("remaining = %d, want 1800", hb.RemainingSeconds)
	}
	if hb.IsExpired {
		t.Error("belum expired")
	}
	if !hb.ServerTime.Equal(base) {
		t.Error("server_time harus dari jam injeksi")
	}

	// lewat waktu
	f.now = base.Add(time.Hour)
	_, hb, _ = f.uc.Heartbeat(ctxBg(), f.userID, f.attemptID)
	if !hb.IsExpired || hb.RemainingSeconds != 0 {
		t.Errorf("expired=%v remaining=%d", hb.IsExpired, hb.RemainingSeconds)
	}
}

func TestAttemptAutosave_SkipsForeignQuestions(t *testing.T) {
	f := newAttemptFixture(time.Now())
	in1, in2 := uuid.New(), uuid.New()
	f.sections.inExam[in1.String()] = true // in2 di luar ujian

	saved, err := f.uc.Autosave(ctxBg(), f.userID, f.attemptID, []AutosaveItem{
		{QuestionID: in1, Value: "B"},
		{QuestionID: in2, Value: "C"},
	})
	if err != nil {
		t.Fatalf("autosave: %v", err)
	}
	if saved != 1 {
		t.Errorf("saved = %d, want 1", saved)
	}
	if len(f.answers.upserted) != 1 || f.answers.upserted[0].qid != in1 {
		t.Errorf("yang tersimpan salah: %+v", f.answers.upserted)
	}
}

func TestAttemptSubmit(t *testing.T) {
	t.Run("tanpa confirm → 422", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		_, err := f.uc.Submit(ctxBg(), f.userID, f.attemptID, false)
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("idempotent: sudah submitted → tanpa FinalizeSubmit baru", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		f.attempts.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
			a := f.currentAttempt()
			a.Status = model.AttemptStatusSubmitted
			return a, nil
		}
		out, err := f.uc.Submit(ctxBg(), f.userID, f.attemptID, true)
		if err != nil {
			t.Fatalf("submit: %v", err)
		}
		if out.Status != model.AttemptStatusSubmitted {
			t.Errorf("status = %q", out.Status)
		}
		if len(f.attempts.finalized) != 0 {
			t.Error("submitted ulang tidak boleh FinalizeSubmit lagi")
		}
	})

	t.Run("in_progress lewat waktu → tetap dikumpulkan sbg submitted", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		f.now = f.now.Add(time.Hour)
		f.attempts.findByIDFn = func(_ context.Context, _ uuid.UUID) (*model.ExamAttempt, error) {
			a := f.currentAttempt()
			if len(f.attempts.finalized) > 0 { // repo stateful: sudah difinalisasi
				a.Status = model.AttemptStatusSubmitted
				a.SubmittedAt = &f.now
			}
			return a, nil
		}
		out, err := f.uc.Submit(ctxBg(), f.userID, f.attemptID, true)
		if err != nil {
			t.Fatalf("submit: %v", err)
		}
		if out.Status != model.AttemptStatusSubmitted {
			t.Errorf("status = %q", out.Status)
		}
		if len(f.attempts.marked) != 1 {
			t.Error("harus ditandai expired dulu")
		}
		if len(f.attempts.finalized) != 1 {
			t.Error("FinalizeSubmit harus dipanggil")
		}
	})

	t.Run("showResultImmediately → penilaian objektif jalan", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		f.exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
			return &model.Exam{BaseModel: model.BaseModel{ID: f.examID}, ShowResultImmediately: true}, nil
		}
		qid := uuid.New()
		f.grading.questionsByID = map[uuid.UUID]*model.Question{
			qid: {BaseModel: model.BaseModel{ID: qid}, QuestionType: model.QuestionTypeMultipleChoice,
				ScoreWeight: 10, Options: []model.Option{{Label: "A", IsCorrect: true}, {Label: "B"}}},
		}
		f.answers.byAttempt[f.attemptID] = []model.ExamAnswer{
			{AttemptID: f.attemptID, QuestionID: qid, AnswerValue: "A"}, // benar
		}
		f.attempts.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
			a := f.currentAttempt()
			a.Score = fltPtr(10)
			if len(f.attempts.finalized) > 0 {
				a.Status = model.AttemptStatusSubmitted
				a.SubmittedAt = &f.now
			}
			return a, nil
		}

		out, err := f.uc.Submit(ctxBg(), f.userID, f.attemptID, true)
		if err != nil {
			t.Fatalf("submit: %v", err)
		}
		if len(f.grading.updatedGrading) != 1 {
			t.Error("UpdateGrading tidak dipanggil utk jawaban objektif")
		}
		if len(f.grading.updatedScore) != 1 || f.grading.updatedScore[0].score != 10 {
			t.Errorf("skor total = %+v, want 10", f.grading.updatedScore)
		}
		if out.Score == nil || *out.Score != 10 {
			t.Errorf("skor attempt = %v", out.Score)
		}
	})
}

func TestAttemptDiscussion_Gating(t *testing.T) {
	t.Run("belum submit → 403", func(t *testing.T) {
		f := newAttemptFixture(time.Now()) // in_progress
		_, _, err := f.uc.GetDiscussion(ctxBg(), f.userID, f.attemptID)
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})

	t.Run("allow_discussion off → 403", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		f.attempts.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
			a := f.currentAttempt()
			a.Status = model.AttemptStatusSubmitted
			return a, nil
		}
		_, _, err := f.uc.GetDiscussion(ctxBg(), f.userID, f.attemptID)
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})

	t.Run("aktif → kunci jawaban & jawaban siswa tampil", func(t *testing.T) {
		base := time.Now()
		f := newAttemptFixture(base)
		f.attempts.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
			a := f.currentAttempt()
			a.Status = model.AttemptStatusSubmitted
			return a, nil
		}
		f.exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
			return &model.Exam{BaseModel: model.BaseModel{ID: f.examID}, AllowDiscussion: true}, nil
		}
		qid := uuid.New()
		f.sections.listExamQFn = func(context.Context, *model.Exam) ([]repository.ExamQuestionGroup, error) {
			return []repository.ExamQuestionGroup{{
				Section:   model.ExamSection{Name: "PG", Sequence: 1},
				Questions: []model.Question{{BaseModel: model.BaseModel{ID: qid}, QuestionType: model.QuestionTypeMultipleChoice, Content: "2+2?", Options: []model.Option{{Label: "A", Content: "3"}, {Label: "B", Content: "4", IsCorrect: true}}}},
			}}, nil
		}
		f.answers.byAttempt[f.attemptID] = []model.ExamAnswer{
			{AttemptID: f.attemptID, QuestionID: qid, AnswerValue: "A", IsFlagged: true},
		}

		_, items, err := f.uc.GetDiscussion(ctxBg(), f.userID, f.attemptID)
		if err != nil {
			t.Fatalf("discussion: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("items = %d", len(items))
		}
		it := items[0]
		if len(it.CorrectKeys) != 1 || it.CorrectKeys[0] != "B" {
			t.Errorf("correct keys = %v", it.CorrectKeys)
		}
		if it.AnswerValue != "A" || !it.IsFlagged {
			t.Errorf("jawaban siswa tidak tampil: %+v", it)
		}
	})
}

// ---------- unit murni: gradeObjective ----------

func mcQuestion() *model.Question {
	return &model.Question{
		QuestionType: model.QuestionTypeMultipleChoice, ScoreWeight: 10,
		Options: []model.Option{{Label: "A"}, {Label: "B", IsCorrect: true}},
	}
}

func TestGradeObjective_MC(t *testing.T) {
	q := mcQuestion()
	score, isCorrect, auto := gradeObjective(q, "B", 0)
	if !auto || !*isCorrect || score != 10 {
		t.Errorf("jawaban benar: score=%v isCorrect=%v auto=%v", score, *isCorrect, auto)
	}
	score, isCorrect, _ = gradeObjective(q, "b", 0) // case-insensitive terhadap kunci "B"
	if !*isCorrect {
		t.Error("huruf kecil harus tetap benar")
	}
	score, isCorrect, _ = gradeObjective(q, "A", 2)
	if *isCorrect || score != -2 {
		t.Errorf("salah + negatif: score=%v want -2", score)
	}
	score, _, _ = gradeObjective(q, "", 2)
	if score != 0 {
		t.Errorf("kosong tidak kena negatif, got %v", score)
	}
}

func TestGradeObjective_MultiAnswer(t *testing.T) {
	q := &model.Question{
		QuestionType: model.QuestionTypeMultipleAnswer, ScoreWeight: 6,
		Options: []model.Option{{Label: "A", IsCorrect: true}, {Label: "B", IsCorrect: true}, {Label: "C"}},
	}
	score, isCorrect, auto := gradeObjective(q, "a,b", 0)
	if !auto || !*isCorrect || score != 6 {
		t.Errorf("multi lengkap: score=%v correct=%v auto=%v", score, *isCorrect, auto)
	}
	_, isCorrect, _ = gradeObjective(q, "a", 0) // parsial
	if *isCorrect {
		t.Error("jawaban parsial harus salah")
	}
	_, isCorrect, _ = gradeObjective(q, "a,c", 0) // ada yang salah
	if *isCorrect {
		t.Error("ada opsi salah harus salah")
	}
}

func TestGradeObjective_ShortAnswer(t *testing.T) {
	keys := "jakarta\nJakarta Utara"
	q := &model.Question{QuestionType: model.QuestionTypeShortAnswer, ScoreWeight: 5, AnswerKeys: &keys}
	_, isCorrect, auto := gradeObjective(q, "  jakarta ", 0)
	if !auto || !*isCorrect {
		t.Errorf("isian cocok: correct=%v auto=%v", *isCorrect, auto)
	}
	_, isCorrect, _ = gradeObjective(q, "bandung", 0)
	if *isCorrect {
		t.Error("isian salah harus salah")
	}
}

func TestGradeObjective_EssayNotAuto(t *testing.T) {
	q := &model.Question{QuestionType: model.QuestionTypeEssay, ScoreWeight: 25}
	_, isCorrect, auto := gradeObjective(q, "apapun", 0)
	if auto || isCorrect != nil {
		t.Errorf("esai tidak boleh dinilai otomatis: auto=%v isCorrect=%v", auto, isCorrect)
	}
}

func TestGradeObjective_TrueFalse(t *testing.T) {
	q := &model.Question{
		QuestionType: model.QuestionTypeTrueFalse, ScoreWeight: 4,
		Options: []model.Option{{Label: "BENAR", IsCorrect: true}, {Label: "SALAH"}},
	}
	_, isCorrect, _ := gradeObjective(q, "benar", 0)
	if !*isCorrect {
		t.Error("TF harus case-insensitive")
	}
	_, isCorrect, _ = gradeObjective(q, "SALAH", 0)
	if *isCorrect {
		t.Error("jawaban berlawanan harus salah")
	}
}
