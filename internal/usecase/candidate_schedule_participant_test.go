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

// ---------- CandidateExamUsecase ----------

type candidateFixture struct {
	uc           *CandidateExamUsecase
	userID       uuid.UUID
	studentID    uuid.UUID
	examID       uuid.UUID
	attempts     *fakeAttemptRepo
	students     *fakeStudentRepo
	schedules    *fakeScheduleRepo
	exams        *fakeExamRepo
	participants *fakeParticipantRepo
	now          time.Time
}

func newCandidateFixture() *candidateFixture {
	userID, studentID, examID := uuid.New(), uuid.New(), uuid.New()
	attempts := &fakeAttemptRepo{}
	students := &fakeStudentRepo{}
	schedules := &fakeScheduleRepo{findByExam: map[uuid.UUID]*model.ExamSchedule{}}
	exams := &fakeExamRepo{}
	participants := &fakeParticipantRepo{}

	now := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	f := &candidateFixture{
		userID: userID, studentID: studentID, examID: examID,
		attempts: attempts, students: students, schedules: schedules,
		exams: exams, participants: participants, now: now,
	}

	students.findByUID = func(_ context.Context, uid uuid.UUID) (*model.Student, error) {
		if uid == userID {
			return &model.Student{BaseModel: model.BaseModel{ID: studentID}}, nil
		}
		return nil, notFound()
	}
	exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
		if id == examID {
			return &model.Exam{
				BaseModel: model.BaseModel{ID: examID},
				Status:    model.ExamStatusPublished, MaxAttempts: 2, DurationMinutes: 60,
			}, nil
		}
		return nil, notFound()
	}
	participants.assignedOverride = map[uuid.UUID]bool{studentID: true}
	schedules.findByExam[examID] = &model.ExamSchedule{
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), Token: "RAHASIA",
	}

	f.uc = NewCandidateExamUsecase(attempts, students, schedules, exams, participants)
	f.uc.now = func() time.Time { return f.now }
	return f
}

func TestCandidateListExams(t *testing.T) {
	f := newCandidateFixture()
	f.attempts.candExams = []repository.CandidateExamRow{{Title: "Ujian A", IsAssigned: true}}

	rows, err := f.uc.ListExams(ctxBg(), f.userID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 || rows[0].Title != "Ujian A" {
		t.Errorf("rows = %+v", rows)
	}

	// bukan siswa → 403
	if _, err := f.uc.ListExams(ctxBg(), uuid.New()); apperror.From(err).Code != 403 {
		t.Errorf("non-siswa want 403, got %v", err)
	}
}

func TestCandidateLoadContext_Gates(t *testing.T) {
	t.Run("ujian belum published → 403", func(t *testing.T) {
		f := newCandidateFixture()
		f.exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
			return &model.Exam{BaseModel: model.BaseModel{ID: f.examID}, Status: model.ExamStatusDraft}, nil
		}
		err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "")
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})

	t.Run("bukan peserta → 403", func(t *testing.T) {
		f := newCandidateFixture()
		f.participants.assignedOverride = map[uuid.UUID]bool{} // tidak terdaftar
		err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "")
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})
}

func TestCandidateValidateToken(t *testing.T) {
	t.Run("belum ada jadwal → 422", func(t *testing.T) {
		f := newCandidateFixture()
		f.schedules.findByExam = map[uuid.UUID]*model.ExamSchedule{}
		err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "")
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("sebelum mulai → 403", func(t *testing.T) {
		f := newCandidateFixture()
		f.now = f.now.Add(-2 * time.Hour)
		err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "")
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})

	t.Run("token salah saat tokenEnabled → 403", func(t *testing.T) {
		f := newCandidateFixture()
		f.exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
			return &model.Exam{BaseModel: model.BaseModel{ID: f.examID}, Status: model.ExamStatusPublished, TokenEnabled: true, MaxAttempts: 2}, nil
		}
		err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "SALAH")
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
		// token benar (case-insensitive) + kuota tersisa → lolos
		if err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "rahasia"); err != nil {
			t.Errorf("token benar harus lolos: %v", err)
		}
	})

	t.Run("kuota habis → 403", func(t *testing.T) {
		f := newCandidateFixture()
		f.exams.findByIDFn = func(_ context.Context, id uuid.UUID) (*model.Exam, error) {
			return &model.Exam{BaseModel: model.BaseModel{ID: f.examID}, Status: model.ExamStatusPublished, MaxAttempts: 2}, nil
		}
		f.attempts.counts = map[string]int64{key2(f.examID, f.studentID): 2}
		err := f.uc.ValidateToken(ctxBg(), f.userID, f.examID, "")
		if ae := apperror.From(err); ae.Code != 403 {
			t.Fatalf("want 403, got %v", err)
		}
	})
}

func TestCandidateStart(t *testing.T) {
	t.Run("resume attempt aktif tanpa buat baru", func(t *testing.T) {
		f := newCandidateFixture()
		active := &model.ExamAttempt{BaseModel: model.BaseModel{ID: uuid.New()}, ExamID: f.examID, StudentID: f.studentID}
		f.attempts.findActive = map[string]*model.ExamAttempt{key2(f.examID, f.studentID): active}

		out, err := f.uc.Start(ctxBg(), f.userID, f.examID, "")
		if err != nil {
			t.Fatalf("start: %v", err)
		}
		if out.ID != active.ID {
			t.Error("harus resume attempt aktif")
		}
		if len(f.attempts.created) != 0 {
			t.Error("tidak boleh buat attempt baru")
		}
	})

	t.Run("attempt baru: nomor, status, expiry dari durasi", func(t *testing.T) {
		f := newCandidateFixture()
		out, err := f.uc.Start(ctxBg(), f.userID, f.examID, "")
		if err != nil {
			t.Fatalf("start: %v", err)
		}
		if len(f.attempts.created) != 1 {
			t.Fatal("attempt tidak dibuat")
		}
		a := f.attempts.created[0]
		if a.AttemptNo != 1 || a.Status != model.AttemptStatusInProgress {
			t.Errorf("attempt = no %d status %q", a.AttemptNo, a.Status)
		}
		wantExpiry := f.now.Add(60 * time.Minute)
		if !a.ExpiresAt.Equal(wantExpiry) {
			t.Errorf("expiresAt = %v, want %v", a.ExpiresAt, wantExpiry)
		}
		if out != a {
			t.Error("harus kembalikan attempt yang dibuat")
		}
	})

	t.Run("attempt kedua: nomor naik", func(t *testing.T) {
		f := newCandidateFixture()
		f.attempts.counts = map[string]int64{key2(f.examID, f.studentID): 1}
		if _, err := f.uc.Start(ctxBg(), f.userID, f.examID, ""); err != nil {
			t.Fatalf("start: %v", err)
		}
		if f.attempts.created[0].AttemptNo != 2 {
			t.Errorf("attempt no = %d, want 2", f.attempts.created[0].AttemptNo)
		}
	})
}

// ---------- ExamScheduleUsecase ----------

func newScheduleUC() (*ExamScheduleUsecase, *fakeScheduleRepo, *fakeExamRepo) {
	schedules := &fakeScheduleRepo{findByExam: map[uuid.UUID]*model.ExamSchedule{}}
	exams := &fakeExamRepo{}
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{}, nil
	}
	return NewExamScheduleUsecase(schedules, exams), schedules, exams
}

func TestScheduleCreate_Validation(t *testing.T) {
	uc, _, _ := newScheduleUC()
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)

	t.Run("end <= start → 422", func(t *testing.T) {
		_, err := uc.Create(ctxBg(), uuid.New(), ExamScheduleInput{StartTime: base, EndTime: base})
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("token duplikat → 422", func(t *testing.T) {
		uc, schedules, _ := newScheduleUC()
		schedules.tokenTaken = map[string]bool{"DIPAKAI": true}
		_, err := uc.Create(ctxBg(), uuid.New(), ExamScheduleInput{
			StartTime: base, EndTime: base.Add(time.Hour), Token: "DIPAKAI",
		})
		if ae := apperror.From(err); ae.Code != 409 {
			t.Fatalf("token dipakai want 409, got %v", err)
		}
	})

	t.Run("valid → tersimpan", func(t *testing.T) {
		uc, _, _ := newScheduleUC()
		out, err := uc.Create(ctxBg(), uuid.New(), ExamScheduleInput{
			StartTime: base, EndTime: base.Add(2 * time.Hour), Token: "UNIK01",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if out.Token != "UNIK01" {
			t.Errorf("token = %q", out.Token)
		}
	})
}

// ---------- ExamParticipantUsecase ----------

func newParticipantUC() (*ExamParticipantUsecase, *fakeParticipantRepo, *fakeExamRepo, *fakeStudentRepo, *fakeClassRepo) {
	participants := &fakeParticipantRepo{}
	exams := &fakeExamRepo{}
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{}, nil
	}
	students := &fakeStudentRepo{}
	classes := &fakeClassRepo{}
	return NewExamParticipantUsecase(participants, exams, classes, students), participants, exams, students, classes
}

func TestParticipantAssignClasses(t *testing.T) {
	t.Run("kelas tak ada → 422", func(t *testing.T) {
		uc, _, _, _, classes := newParticipantUC()
		classes.existsFn = func(context.Context, uuid.UUID) (bool, error) { return false, nil }

		_, err := uc.AssignClasses(ctxBg(), uuid.New(), []uuid.UUID{uuid.New()})
		ae := apperror.From(err)
		if ae.Code != 422 || ae.Details["class_ids"] != "kelas tidak ditemukan" {
			t.Fatalf("want 422 class_ids, got %v", err)
		}
	})

	t.Run("siswa dari kelas di-assign", func(t *testing.T) {
		uc, participants, _, _, classes := newParticipantUC()
		classes.existsFn = func(context.Context, uuid.UUID) (bool, error) { return true, nil }
		sid := uuid.New()
		participants.studentsByClassesFn = func(context.Context, []uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{sid}, nil
		}

		res, err := uc.AssignClasses(ctxBg(), uuid.New(), []uuid.UUID{uuid.New()})
		if err != nil {
			t.Fatalf("assign: %v", err)
		}
		if res == nil || res.Assigned != 1 {
			t.Errorf("assigned = %+v, want 1", res)
		}
	})
}

func TestParticipantRemove(t *testing.T) {
	uc, participants, _, _, _ := newParticipantUC()
	pid, sid := uuid.New(), uuid.New()
	examID := uuid.New()
	participants.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamParticipant, error) {
		return &model.ExamParticipant{BaseModel: model.BaseModel{ID: pid}, ExamID: examID, StudentID: sid}, nil
	}

	if err := uc.Remove(ctxBg(), examID, pid); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(participants.removed) != 1 || participants.removed[0] != pid {
		t.Error("RemoveWithCleanup tidak dipanggil")
	}
}
