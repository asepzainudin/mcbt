package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

// ---------- Dashboard ----------

func TestDashboardAdmin(t *testing.T) {
	dash := &fakeDashRepo{adminStats: &repository.AdminStats{
		TotalStudents: 12, TotalTeachers: 3, TotalQuestionBanks: 5, PublishedExams: 2, OngoingExams: 1, TotalAttempts: 40,
	}}
	students := &fakeStudentRepo{}
	teachers := &fakeTeacherRepo{}
	uc := NewDashboardUsecase(dash, students, teachers)

	stats, err := uc.Admin(ctxBg())
	if err != nil {
		t.Fatalf("admin dashboard: %v", err)
	}
	if stats.TotalStudents != 12 || stats.PublishedExams != 2 {
		t.Errorf("stats salah: %+v", stats)
	}
}

func TestDashboardTeacher(t *testing.T) {
	dash := &fakeDashRepo{}
	students := &fakeStudentRepo{}
	teachers := &fakeTeacherRepo{}
	userID := uuid.New()
	teachers.findByUID = func(context.Context, uuid.UUID) (*model.Teacher, error) {
		return &model.Teacher{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID}, nil
	}
	dash.teacherStats = &repository.TeacherStats{TotalBanks: 4, TotalQuestions: 25, PublishedExams: 1}

	uc := NewDashboardUsecase(dash, students, teachers)
	stats, err := uc.Teacher(ctxBg(), userID)
	if err != nil {
		t.Fatalf("teacher dashboard: %v", err)
	}
	if stats.TotalBanks != 4 || stats.TotalQuestions != 25 {
		t.Errorf("stats salah: %+v", stats)
	}

	// bukan guru → 404
	teachers.findByUID = func(context.Context, uuid.UUID) (*model.Teacher, error) { return nil, notFound() }
	if _, err := uc.Teacher(ctxBg(), uuid.New()); apperror.From(err).Code != 404 {
		t.Errorf("want 404, got %v", err)
	}
}

func TestDashboardStudent(t *testing.T) {
	dash := &fakeDashRepo{}
	students := &fakeStudentRepo{}
	teachers := &fakeTeacherRepo{}
	userID := uuid.New()
	students.findByUID = func(context.Context, uuid.UUID) (*model.Student, error) {
		return &model.Student{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID}, nil
	}
	avg, best := 82.5, 95.0
	dash.studentStats = &repository.StudentStats{
		AssignedExams: 5, CompletedExams: 3, PassedExams: 2,
		AverageScore: sql.NullFloat64{Float64: avg, Valid: true}, BestScore: sql.NullFloat64{Float64: best, Valid: true},
	}

	uc := NewDashboardUsecase(dash, students, teachers)
	out, err := uc.Student(ctxBg(), userID)
	if err != nil {
		t.Fatalf("student dashboard: %v", err)
	}
	if out.AssignedExams != 5 || out.CompletedExams != 3 || out.PassedExams != 2 {
		t.Errorf("hitungan salah: %+v", out)
	}
	if out.AverageScore == nil || *out.AverageScore != 82.5 {
		t.Errorf("rata-rata = %v", out.AverageScore)
	}
	if out.BestScore == nil || *out.BestScore != 95 {
		t.Errorf("terbaik = %v", out.BestScore)
	}
}

// ---------- Profile ----------

func TestProfileUpdate(t *testing.T) {
	t.Run("nama & telepon tersimpan", func(t *testing.T) {
		profiles := &fakeProfileRepo{}
		users := &fakeUserRepo{}
		uc := NewProfileUsecase(profiles, users)

		phone := "08123456789"
		out, err := uc.Update(ctxBg(), uuid.New(), UpdateProfileInput{Name: "Nama Baru", Phone: &phone})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if profiles.nameUpdate != "Nama Baru" {
			t.Errorf("nama = %q", profiles.nameUpdate)
		}
		if out == nil {
			t.Error("profil terbaru harus dikembalikan")
		}
	})

	t.Run("user tak ada → 404", func(t *testing.T) {
		profiles := &fakeProfileRepo{}
		users := &fakeUserRepo{}
		users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) { return nil, notFound() }
		uc := NewProfileUsecase(profiles, users)

		_, err := uc.Update(ctxBg(), uuid.New(), UpdateProfileInput{Name: "X"})
		if apperror.From(err).Code != 404 {
			t.Errorf("want 404, got %v", err)
		}
	})
}

// ---------- Export ----------

func TestExportExamResults(t *testing.T) {
	ranker := &fakeExamRanker{ranked: []RankedResult{
		{Rank: 1, Nis: "001", StudentName: "Satu", Score: fltPtr(90), Passed: true, PassingGrade: 70},
		{Rank: 2, Nis: "002", StudentName: "Dua", Score: fltPtr(60), PassingGrade: 70},
	}}
	exams := &fakeExamRepo{}
	examID := uuid.New()
	exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) {
		return &model.Exam{BaseModel: model.BaseModel{ID: examID}, Title: "Ujian Akhir"}, nil
	}
	students := &fakeStudentRepo{}
	teachers := &fakeTeacherRepo{}
	uc := NewExportUsecase(ranker, exams, students, teachers)

	t.Run("xlsx valid + filename rapi", func(t *testing.T) {
		file, err := uc.ExamResults(ctxBg(), examID, "xlsx")
		if err != nil {
			t.Fatalf("export xlsx: %v", err)
		}
		if file.Filename != "hasil-ujian-ujian-akhir.xlsx" {
			t.Errorf("filename = %q", file.Filename)
		}
		if len(file.Data) == 0 {
			t.Error("data kosong")
		}
	})

	t.Run("pdf valid", func(t *testing.T) {
		file, err := uc.ExamResults(ctxBg(), examID, "pdf")
		if err != nil {
			t.Fatalf("export pdf: %v", err)
		}
		if file.ContentType != "application/pdf" || len(file.Data) == 0 {
			t.Errorf("pdf tidak valid: %s (%d bytes)", file.ContentType, len(file.Data))
		}
	})

	t.Run("format aneh → 400", func(t *testing.T) {
		_, err := uc.ExamResults(ctxBg(), examID, "docx")
		if apperror.From(err).Code != 400 {
			t.Errorf("want 400, got %v", err)
		}
	})

	t.Run("ujian tak ada → 404", func(t *testing.T) {
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		if _, err := uc.ExamResults(ctxBg(), uuid.New(), "xlsx"); apperror.From(err).Code != 404 {
			t.Errorf("want 404, got %v", err)
		}
	})
}

// ---------- Role ----------

func TestRoleAssign(t *testing.T) {
	roles := &fakeRoleRepo{}
	users := &fakeUserRepo{}
	uc := NewRoleUsecase(roles, users)

	userID := uuid.New()
	roleA, roleB := uuid.New(), uuid.New()
	if err := uc.AssignToUser(ctxBg(), userID, []uuid.UUID{roleA, roleB}); err != nil {
		t.Fatalf("assign: %v", err)
	}
	got := roles.replaced[userID]
	if len(got) != 2 {
		t.Errorf("roles = %v, want 2", got)
	}
}

// ---------- Master: Academic Year ----------

func TestAcademicYearActivate(t *testing.T) {
	repo := &fakeAYRepo{}
	uc := NewAcademicYearUsecase(repo)
	id := uuid.New()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.AcademicYear, error) {
		return &model.AcademicYear{BaseModel: model.BaseModel{ID: id}}, nil
	}

	if _, err := uc.Activate(ctxBg(), id); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(repo.activateCalls) != 1 || repo.activateCalls[0] != id {
		t.Errorf("activate tidak diteruskan: %v", repo.activateCalls)
	}
}

func TestAcademicYearCreate_Duplicate(t *testing.T) {
	repo := &fakeAYRepo{}
	uc := NewAcademicYearUsecase(repo)
	repo.dupFn = func(context.Context, string, string, *uuid.UUID) (bool, error) { return true, nil }

	_, err := uc.Create(ctxBg(), AcademicYearInput{Year: "2026/2027", Semester: "ganjil"})
	if ae := apperror.From(err); ae.Code != 409 {
		t.Fatalf("want 409, got %v", err)
	}
}

// ---------- Master: Subject ----------

func TestSubjectCreate_Duplicate(t *testing.T) {
	subjects := &fakeSubjectRepo{dupFn: func(context.Context, string, *uuid.UUID) (bool, error) {
		return true, nil
	}}
	uc := NewSubjectUsecase(subjects)

	_, err := uc.Create(ctxBg(), SubjectInput{Code: "MATH", Name: "Matematika"})
	if apperror.From(err).Code != 409 {
		t.Errorf("want 409, got %v", err)
	}
}

func TestSubjectDelete(t *testing.T) {
	subjects := &fakeSubjectRepo{deleteErr: notFound()}
	uc := NewSubjectUsecase(subjects)

	err := uc.Delete(ctxBg(), uuid.New())
	if apperror.From(err).Code != 404 {
		t.Errorf("want 404, got %v", err)
	}
}

// ---------- Tambahan coverage: operasi CRUD & utilitas sederhana ----------

func TestSimpleCRUDCoverage(t *testing.T) {
	t.Run("exam Get/Delete notfound & Update", func(t *testing.T) {
		uc, repo, subjects, _, _, _ := newExamUC()
		subjectOK(subjects)
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return nil, notFound() }
		if _, err := uc.Get(ctxBg(), uuid.New()); apperror.From(err).Code != 404 {
			t.Error("Get notfound")
		}
		repo.findByIDFn = func(_ context.Context, _ uuid.UUID) (*model.Exam, error) {
			e := model.Exam{BaseModel: model.BaseModel{ID: uuid.New()}, Title: "Lama"}
			return &e, nil
		}
		if _, err := uc.Update(ctxBg(), uuid.New(), ExamInput{Title: "Baru", SubjectID: uuid.New()}); err != nil {
			t.Errorf("update: %v", err)
		}
	})

	t.Run("bank Archive & List", func(t *testing.T) {
		uc, repo, _, _ := newBankUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.QuestionBank, error) {
			return bankOf(nil), nil
		}
		if _, err := uc.Archive(ctxBg(), uuid.New()); err != nil {
			t.Errorf("archive: %v", err)
		}
		if repo.setStatus[0].status != model.BankStatusArchived {
			t.Error("status archived salah")
		}
		if _, _, err := uc.List(ctxBg(), "", nil, nil, 1, 10); err != nil {
			t.Errorf("list: %v", err)
		}
	})

	t.Run("section delete/remove/list/examQuestions", func(t *testing.T) {
		uc, sections, exams, _ := newSectionUC()
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return &model.Exam{}, nil }
		sections.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSection, error) {
			return &model.ExamSection{}, nil
		}
		sections.deleteErr = gorm.ErrRecordNotFound
		if apperror.From(uc.Delete(ctxBg(), uuid.New())).Code != 404 {
			t.Error("delete notfound")
		}
		sections.deleteErr = nil
		if err := uc.Delete(ctxBg(), uuid.New()); err != nil {
			t.Errorf("delete ok: %v", err)
		}
		sections.removeErr = gorm.ErrRecordNotFound
		if apperror.From(uc.RemoveQuestion(ctxBg(), uuid.New(), uuid.New())).Code != 404 {
			t.Error("remove notfound")
		}
		if _, err := uc.ListQuestions(ctxBg(), uuid.New()); err != nil {
			t.Errorf("listQuestions: %v", err)
		}
		if _, err := uc.ExamQuestions(ctxBg(), uuid.New()); err != nil {
			t.Errorf("examQuestions: %v", err)
		}
		if _, err := uc.ListByExam(ctxBg(), uuid.New()); err != nil {
			t.Errorf("listByExam: %v", err)
		}
	})

	t.Run("attempt GetQuestions & SetFlag", func(t *testing.T) {
		f := newAttemptFixture(time.Now())
		f.sections.listExamQFn = func(context.Context, *model.Exam) ([]repository.ExamQuestionGroup, error) {
			return []repository.ExamQuestionGroup{}, nil
		}
		if _, _, err := f.uc.GetQuestions(ctxBg(), f.userID, f.attemptID); err != nil {
			t.Errorf("getQuestions: %v", err)
		}
		qid := uuid.New()
		f.sections.inExam[qid.String()] = true
		if _, err := f.uc.SetFlag(ctxBg(), f.userID, f.attemptID, qid, true); err != nil {
			t.Errorf("setFlag: %v", err)
		}
		if len(f.answers.flags) != 1 || !f.answers.flags[0].flagged {
			t.Error("SetFlag tidak diteruskan")
		}
	})

	t.Run("schedule GetByExam/Delete/Update/GenerateToken", func(t *testing.T) {
		uc, schedules, _ := newScheduleUC()
		examID := uuid.New()
		schedules.findByExam[examID] = &model.ExamSchedule{Token: "TOK"}
		if _, err := uc.GetByExam(ctxBg(), examID); err != nil {
			t.Errorf("getByExam: %v", err)
		}
		if err := uc.Delete(ctxBg(), uuid.New()); err != nil {
			t.Errorf("delete: %v", err)
		}
		base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
		sid := uuid.New()
		schedules.findByIDFn = func(context.Context, uuid.UUID) (*model.ExamSchedule, error) {
			return &model.ExamSchedule{BaseModel: model.BaseModel{ID: sid}, StartTime: base, EndTime: base.Add(time.Hour)}, nil
		}
		if _, err := uc.Update(ctxBg(), sid, ExamScheduleInput{StartTime: base, EndTime: base.Add(2 * time.Hour), Token: "BARU1"}); err != nil {
			t.Errorf("update: %v", err)
		}
		if _, _, err := uc.GenerateToken(ctxBg(), sid); err != nil {
			t.Errorf("generateToken: %v", err)
		}
	})

	t.Run("participant List & AssignIndividuals", func(t *testing.T) {
		uc, participants, _, _, _ := newParticipantUC()
		examID := uuid.New()
		participants.listByExam = map[uuid.UUID][]model.ExamParticipant{
			examID: {{BaseModel: model.BaseModel{ID: uuid.New()}, ExamID: examID}},
		}
		if out, err := uc.List(ctxBg(), examID); err != nil || len(out) != 1 {
			t.Errorf("list: %v (%d)", err, len(out))
		}
		res, err := uc.AssignIndividuals(ctxBg(), examID, []uuid.UUID{uuid.New()})
		if err != nil || res.Assigned != 1 {
			t.Errorf("assign individuals: %v (%+v)", err, res)
		}
		// siswa tidak ada → 422
		participants.studentsExistFn = func(context.Context, []uuid.UUID) (bool, error) { return false, nil }
		if _, err := uc.AssignIndividuals(ctxBg(), examID, []uuid.UUID{uuid.New()}); apperror.From(err).Code != 422 {
			t.Error("want 422")
		}
	})

	t.Run("grading UngradedEssays & ExamGradingSheet", func(t *testing.T) {
		uc, grading, exams := newGradingUC()
		examID := uuid.New()
		exams.findByIDFn = func(context.Context, uuid.UUID) (*model.Exam, error) { return &model.Exam{}, nil }
		grading.ungraded = []repository.UngradedEssayRow{}
		if _, err := uc.UngradedEssays(ctxBg(), examID); err != nil {
			t.Errorf("ungraded: %v", err)
		}
		grading.submitted = []model.ExamAttempt{{BaseModel: model.BaseModel{ID: uuid.New()}, Status: model.AttemptStatusSubmitted, Score: fltPtr(5)}}
		if _, err := uc.ExamGradingSheet(ctxBg(), examID); err != nil {
			t.Errorf("sheet: %v", err)
		}
	})

	t.Run("export Students & Teachers", func(t *testing.T) {
		ranker := &fakeExamRanker{}
		exams := &fakeExamRepo{}
		students := &fakeStudentRepo{}
		teachers := &fakeTeacherRepo{}
		uc := NewExportUsecase(ranker, exams, students, teachers)
		if file, err := uc.Students(ctxBg(), "xlsx"); err != nil || len(file.Data) == 0 {
			t.Errorf("students xlsx: %v", err)
		}
		if file, err := uc.Teachers(ctxBg(), "pdf"); err != nil || len(file.Data) == 0 {
			t.Errorf("teachers pdf: %v", err)
		}
	})

	t.Run("role List & user exists", func(t *testing.T) {
		roles := &fakeRoleRepo{}
		users := &fakeUserRepo{}
		uc := NewRoleUsecase(roles, users)
		if _, total, err := uc.List(ctxBg(), 1, 10); err != nil || total != 0 {
			t.Errorf("list: %v", err)
		}
	})

	t.Run("master AY/Class/Subject CRUD ringkas", func(t *testing.T) {
		ayRepo := &fakeAYRepo{}
		ayUC := NewAcademicYearUsecase(ayRepo)
		ayRepo.findByIDFn = func(context.Context, uuid.UUID) (*model.AcademicYear, error) {
			return &model.AcademicYear{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		}
		if _, err := ayUC.Update(ctxBg(), uuid.New(), AcademicYearInput{Year: "2026/2027", Semester: "genap"}); err != nil {
			t.Errorf("ay update: %v", err)
		}
		if err := ayUC.Delete(ctxBg(), uuid.New()); err != nil {
			t.Errorf("ay delete: %v", err)
		}

		classRepo := &fakeClassRepo{}
		ayRepo2 := &fakeAYRepo{}
		classUC := NewClassUsecase(classRepo, ayRepo2)
		ayID := uuid.New()
		ayRepo2.findByIDFn = func(context.Context, uuid.UUID) (*model.AcademicYear, error) {
			return &model.AcademicYear{BaseModel: model.BaseModel{ID: ayID}}, nil
		}
		if _, err := classUC.Create(ctxBg(), ClassInput{Name: "VII-A", AcademicYearID: ayID}); err != nil {
			t.Errorf("class create: %v", err)
		}

		subjectUC := NewSubjectUsecase(&fakeSubjectRepo{})
		if _, err := subjectUC.Create(ctxBg(), SubjectInput{Code: "BIO", Name: "Biologi"}); err != nil {
			t.Errorf("subject create: %v", err)
		}
	})

	t.Run("student List & Get found", func(t *testing.T) {
		uc, repo, _, _ := newStudentUC()
		repo.listFn = func(context.Context, string, *uuid.UUID, int, int) (*repository.PageResult[model.Student], error) {
			return pageOf([]model.Student{*studentWithUser("A")}), nil
		}
		items, total, err := uc.List(ctxBg(), "", nil, 1, 10)
		if err != nil || total != 1 || len(items) != 1 {
			t.Errorf("list: %v (%d/%d)", err, len(items), total)
		}
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) {
			return studentWithUser("B"), nil
		}
		if _, err := uc.Get(ctxBg(), uuid.New()); err != nil {
			t.Errorf("get: %v", err)
		}
	})
}

func mustErr2(err error) error {
	return mustErr(err)
}
