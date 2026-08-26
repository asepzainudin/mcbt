package server

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/config"
	"github.com/asepzainudin14/mcbt/internal/delivery/http/handler"
	jwtmanager "github.com/asepzainudin14/mcbt/internal/pkg/jwt"
	"github.com/asepzainudin14/mcbt/internal/pkg/storage"
	"github.com/asepzainudin14/mcbt/internal/repository"
	middleware "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type RouterDeps struct {
	Cfg *config.Config
	Log *slog.Logger
	DB  *gorm.DB
}

func NewRouter(d RouterDeps) (*gin.Engine, error) {
	if d.Cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.MaxMultipartMemory = 8 << 20

	users := repository.NewUserRepository(d.DB)
	roles := repository.NewRoleRepository(d.DB)

	tokens := jwtmanager.NewManager(
		d.Cfg.JWTSecret,
		d.Cfg.JWTAccessTTL,
		d.Cfg.JWTRefreshTTL,
	)
	authUC := usecase.NewAuthUsecase(users, tokens)
	roleUC := usecase.NewRoleUsecase(roles, users)
	profileUC := usecase.NewProfileUsecase(repository.NewProfileRepository(d.DB), users)

	ayRepo := repository.NewAcademicYearRepository(d.DB)
	classRepo := repository.NewClassRepository(d.DB)
	subjectRepo := repository.NewSubjectRepository(d.DB)
	examSectionRepo := repository.NewExamSectionRepository(d.DB)
	attemptRepo := repository.NewExamAttemptRepository(d.DB)

	cookies := handler.NewCookieManager(d.Cfg)
	authHandler := handler.NewAuthHandler(authUC, cookies, profileUC)
	roleHandler := handler.NewRoleHandler(roleUC)
	ayHandler := handler.NewAcademicYearHandler(usecase.NewAcademicYearUsecase(ayRepo))
	classHandler := handler.NewClassHandler(usecase.NewClassUsecase(classRepo, ayRepo))
	subjectHandler := handler.NewSubjectHandler(usecase.NewSubjectUsecase(subjectRepo))

	teacherUsecase := usecase.NewTeacherUsecase(
		repository.NewTeacherRepository(d.DB), roles,
	)
	studentUsecase := usecase.NewStudentUsecase(
		repository.NewStudentRepository(d.DB), roles, classRepo,
	)
	teacherHandler := handler.NewTeacherHandler(teacherUsecase)
	studentHandler := handler.NewStudentHandler(studentUsecase)

	studentRepo := repository.NewStudentRepository(d.DB)
	bankRepo := repository.NewQuestionBankRepository(d.DB)
	questionRepo := repository.NewQuestionRepository(d.DB)
	examParticipantRepo := repository.NewExamParticipantRepository(d.DB)
	examScheduleRepo := repository.NewExamScheduleRepository(d.DB)
	_ = examSectionRepo
	mediaRepo := repository.NewMediaRepository(d.DB)

	storageClient, err := storage.NewClient(context.Background(), d.Cfg)
	if err != nil {
		return nil, err
	}

	bankHandler := handler.NewQuestionBankHandler(
		usecase.NewQuestionBankUsecase(bankRepo, subjectRepo, questionRepo),
	)
	questionHandler := handler.NewQuestionHandler(
		usecase.NewQuestionUsecase(questionRepo, bankRepo, examSectionRepo, repository.NewExamAnswerRepository(d.DB)),
	)
	examRepo := repository.NewExamRepository(d.DB)
	accessUC := usecase.NewAccessUsecase(bankRepo, examRepo, examSectionRepo, questionRepo, attemptRepo)
	examHandler := handler.NewExamHandler(
		usecase.NewExamUsecase(examRepo, subjectRepo, ayRepo, bankRepo, attemptRepo),
		accessUC,
	)
	examSectionHandler := handler.NewExamSectionHandler(
		usecase.NewExamSectionUsecase(
			repository.NewExamSectionRepository(d.DB), examRepo, bankRepo,
		),
	)
	gradingHandler := handler.NewGradingHandler(
		usecase.NewGradingUsecase(
			repository.NewGradingRepository(d.DB),
			repository.NewExamAnswerRepository(d.DB),
			attemptRepo,
			examRepo,
		),
	)
	resultHandler := handler.NewResultHandler(
		usecase.NewResultUsecase(
			repository.NewResultRepository(d.DB), examRepo, studentRepo,
		),
		studentRepo,
	)
	examScheduleHandler := handler.NewExamScheduleHandler(
		usecase.NewExamScheduleUsecase(
			repository.NewExamScheduleRepository(d.DB), examRepo,
		),
	)
	examParticipantHandler := handler.NewExamParticipantHandler(
		usecase.NewExamParticipantUsecase(
			repository.NewExamParticipantRepository(d.DB), examRepo, classRepo,
			repository.NewStudentRepository(d.DB),
		),
	)
	attemptEngineUc := usecase.NewAttemptEngineUsecase(
		attemptRepo,
		repository.NewExamAnswerRepository(d.DB),
		studentRepo,
		examSectionRepo,
		examRepo,
		repository.NewGradingRepository(d.DB),
	)
	questionReportUc := usecase.NewQuestionReportUsecase(
		repository.NewQuestionReportRepository(d.DB),
		attemptRepo,
		studentRepo,
		accessUC,
	)
	dashboardUc := usecase.NewDashboardUsecase(
		repository.NewDashboardRepository(d.DB),
		studentRepo,
		repository.NewTeacherRepository(d.DB),
	)
	resultUc := usecase.NewResultUsecase(repository.NewResultRepository(d.DB), examRepo, studentRepo)
	exportUc := usecase.NewExportUsecase(resultUc, examRepo, studentRepo, repository.NewTeacherRepository(d.DB))
	dashboardHandler := handler.NewDashboardHandler(dashboardUc)
	exportHandler := handler.NewExportHandler(exportUc)
	attemptEngineHandler := handler.NewAttemptEngineHandler(attemptEngineUc)
	questionReportHandler := handler.NewQuestionReportHandler(questionReportUc, accessUC)
	candidateExamHandler := handler.NewCandidateExamHandler(
		usecase.NewCandidateExamUsecase(attemptRepo, studentRepo, examScheduleRepo, examRepo, examParticipantRepo),
		attemptEngineUc,
		questionReportUc,
		repository.NewGradingRepository(d.DB),
	)
	questionImportHandler := handler.NewQuestionImportHandler(
		usecase.NewQuestionImportUsecase(usecase.NewImportTokenStore(), questionRepo),
		accessUC,
	)
	mediaHandler := handler.NewMediaHandler(
		usecase.NewMediaUsecase(mediaRepo, storageClient, int64(d.Cfg.MaxUploadMB)<<20),
	)

	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(d.Log))
	r.Use(middleware.Recovery(d.Log))
	r.Use(middleware.ErrorHandler(d.Log))
	r.Use(middleware.CSRFProtection(d.Log))

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh-token", authHandler.RefreshToken)
	}

	v1.GET("/health", handler.NewHealthHandler().Health)

	protected := v1.Group("", middleware.Authenticate(d.Log, tokens, users))
	{
		protected.GET("/auth/me", authHandler.Me)
		protected.GET("/auth/profile", authHandler.Profile)
		protected.PUT("/auth/profile", authHandler.UpdateProfile)
		protected.PUT("/auth/change-password", authHandler.ChangePassword)

		adminOnly := protected.Group("", middleware.RequireRoles("admin"))
		staffOnly := protected.Group("", middleware.RequireRoles("admin", "teacher"))
		{
			// ---------- ADMIN ONLY ----------
			adminOnly.GET("/roles", roleHandler.List)
			adminOnly.POST("/roles/assign", roleHandler.Assign)

			adminOnly.POST("/academic-years", ayHandler.Create)
			adminOnly.PUT("/academic-years/:id", ayHandler.Update)
			adminOnly.PATCH("/academic-years/:id/activate", ayHandler.Activate)
			adminOnly.DELETE("/academic-years/:id", ayHandler.Delete)

			adminOnly.POST("/classes", classHandler.Create)
			adminOnly.PUT("/classes/:id", classHandler.Update)
			adminOnly.DELETE("/classes/:id", classHandler.Delete)

			adminOnly.POST("/subjects", subjectHandler.Create)
			adminOnly.PUT("/subjects/:id", subjectHandler.Update)
			adminOnly.DELETE("/subjects/:id", subjectHandler.Delete)

			adminOnly.GET("/teachers", teacherHandler.List)
			adminOnly.GET("/teachers/import/template", teacherHandler.Template)
			adminOnly.POST("/teachers/import", teacherHandler.Import)
			adminOnly.GET("/teachers/:id", teacherHandler.Get)
			adminOnly.POST("/teachers", teacherHandler.Create)
			adminOnly.PUT("/teachers/:id", teacherHandler.Update)
			adminOnly.POST("/teachers/:id/reset-password", teacherHandler.ResetPassword)
			adminOnly.DELETE("/teachers/:id", teacherHandler.Delete)

			// siswa: tulis hanya admin; lihat di staffOnly
			adminOnly.GET("/students/import/template", studentHandler.Template)
			adminOnly.POST("/students/import", studentHandler.Import)
			adminOnly.POST("/students/:id/change-class", studentHandler.ChangeClass)
			adminOnly.POST("/students/:id/reset-password", studentHandler.ResetPassword)
			adminOnly.POST("/students", studentHandler.Create)
			adminOnly.PUT("/students/:id", studentHandler.Update)
			adminOnly.DELETE("/students/:id", studentHandler.Delete)

			adminOnly.GET("/export/students", exportHandler.Students)
			adminOnly.GET("/export/teachers", exportHandler.Teachers)

			// ---------- STAFF (ADMIN + GURU) ----------
			// master data read-only utk kebutuhan form guru
			staffOnly.GET("/academic-years", ayHandler.List)
			staffOnly.GET("/classes", classHandler.List)
			staffOnly.GET("/subjects", subjectHandler.List)

			// menu siswa: guru read-only
			staffOnly.GET("/students", studentHandler.List)
			staffOnly.GET("/students/:id", studentHandler.Get)

			// media (rich text editor soal)
			staffOnly.POST("/media/upload", mediaHandler.Upload)

			// bank soal: CRUD penuh, scope milik guru
			staffOnly.GET("/question-banks", bankHandler.List)
			staffOnly.POST("/question-banks", bankHandler.Create)
			staffOnly.POST("/question-banks/:id/clone", dataScopeGuard(accessUC, "bank"), bankHandler.Clone)
			staffOnly.PATCH("/question-banks/:id/publish", dataScopeGuard(accessUC, "bank"), bankHandler.Publish)
			staffOnly.PATCH("/question-banks/:id/archive", dataScopeGuard(accessUC, "bank"), bankHandler.Archive)
			staffOnly.PUT("/question-banks/:id", dataScopeGuard(accessUC, "bank"), bankHandler.Update)
			staffOnly.DELETE("/question-banks/:id", dataScopeGuard(accessUC, "bank"), bankHandler.Delete)

			staffOnly.GET("/questions/import/template", questionImportHandler.Template)
			staffOnly.POST("/questions/import/validate", questionImportHandler.Validate)
			staffOnly.POST("/questions/import/process", questionImportHandler.Process)

			staffOnly.GET("/questions", questionHandler.List)
			staffOnly.GET("/questions/:id/preview", questionHandler.Preview)
			staffOnly.GET("/questions/:id", questionHandler.Get)
			staffOnly.POST("/question-banks/:id/questions", dataScopeGuard(accessUC, "bank"), questionHandler.CreateInBank)
			staffOnly.PUT("/questions/:id", dataScopeGuard(accessUC, "question"), questionHandler.Update)
			staffOnly.DELETE("/questions/:id", dataScopeGuard(accessUC, "question"), questionHandler.Delete)
			staffOnly.PUT("/questions/:id/options/reorder", dataScopeGuard(accessUC, "question"), questionHandler.ReorderOptions)
			staffOnly.PUT("/questions/:id/options/:option_id", dataScopeGuard(accessUC, "question"), questionHandler.UpdateOption)

			// ujian: CRUD penuh, scope via bank soal
			staffOnly.GET("/exams", examHandler.List)
			staffOnly.POST("/exams", examHandler.Create)
			staffOnly.GET("/exams/:id", dataScopeGuard(accessUC, "exam"), examHandler.Get)
			staffOnly.PUT("/exams/:id", dataScopeGuard(accessUC, "exam"), examHandler.Update)
			staffOnly.PUT("/exams/:id/settings", dataScopeGuard(accessUC, "exam"), examHandler.UpdateSettings)
			staffOnly.PATCH("/exams/:id/publish", dataScopeGuard(accessUC, "exam"), examHandler.Publish)
			staffOnly.PATCH("/exams/:id/close", dataScopeGuard(accessUC, "exam"), examHandler.Close)
			staffOnly.DELETE("/exams/:id", dataScopeGuard(accessUC, "exam"), examHandler.Delete)

			staffOnly.POST("/exams/:id/calculate-grades", dataScopeGuard(accessUC, "exam"), gradingHandler.CalculateGrades)
			staffOnly.GET("/exams/:id/ungraded-essays", dataScopeGuard(accessUC, "exam"), gradingHandler.UngradedEssays)
			staffOnly.GET("/exams/:id/grading", dataScopeGuard(accessUC, "exam"), gradingHandler.GradingSheet)
			staffOnly.PUT("/attempts/:id/grade-essay", dataScopeGuard(accessUC, "attempt"), gradingHandler.GradeEssay)

			staffOnly.GET("/exams/:id/results", dataScopeGuard(accessUC, "exam"), resultHandler.ExamResults)
			staffOnly.PATCH("/exams/:id/publish-results", dataScopeGuard(accessUC, "exam"), resultHandler.PublishResults)
			staffOnly.GET("/exams/:id/export", dataScopeGuard(accessUC, "exam"), exportHandler.ExamResults)

			staffOnly.GET("/exams/:id/questions", dataScopeGuard(accessUC, "exam"), examSectionHandler.Review)
			staffOnly.GET("/exams/:id/sections", dataScopeGuard(accessUC, "exam"), examSectionHandler.ListByExam)
			staffOnly.POST("/exams/:id/sections", dataScopeGuard(accessUC, "exam"), examSectionHandler.Create)
			staffOnly.PUT("/sections/:id", dataScopeGuard(accessUC, "section"), examSectionHandler.Update)
			staffOnly.DELETE("/sections/:id", dataScopeGuard(accessUC, "section"), examSectionHandler.Delete)
			staffOnly.GET("/sections/:id/questions", dataScopeGuard(accessUC, "section"), examSectionHandler.ListQuestions)
			staffOnly.POST("/sections/:id/questions", dataScopeGuard(accessUC, "section"), examSectionHandler.MapQuestions)
			staffOnly.DELETE("/sections/:id/questions/:question_id", dataScopeGuard(accessUC, "section"), examSectionHandler.RemoveQuestion)

			staffOnly.POST("/exams/:id/schedules", dataScopeGuard(accessUC, "exam"), examScheduleHandler.Create)
			staffOnly.GET("/exams/:id/schedules", dataScopeGuard(accessUC, "exam"), examScheduleHandler.GetByExam)
			staffOnly.PUT("/schedules/:id", dataScopeGuard(accessUC, "exam"), examScheduleHandler.Update)
			staffOnly.DELETE("/schedules/:id", dataScopeGuard(accessUC, "exam"), examScheduleHandler.Delete)
			staffOnly.POST("/schedules/:id/generate-token", dataScopeGuard(accessUC, "exam"), examScheduleHandler.GenerateToken)

			staffOnly.GET("/exams/:id/participants", dataScopeGuard(accessUC, "exam"), examParticipantHandler.List)
			staffOnly.POST("/exams/:id/participants/assign-class", dataScopeGuard(accessUC, "exam"), examParticipantHandler.AssignClass)
			staffOnly.POST("/exams/:id/participants/assign-individual", dataScopeGuard(accessUC, "exam"), examParticipantHandler.AssignIndividual)
			staffOnly.DELETE("/exams/:id/participants/:participant_id", dataScopeGuard(accessUC, "exam"), examParticipantHandler.Remove)

			// laporan soal: guru menangani laporan pada ujiannya sendiri
			staffOnly.GET("/question-reports", questionReportHandler.List)
			staffOnly.PATCH("/question-reports/:id/resolve", questionReportHandler.Resolve)

			// laporan ujian: rekap seluruh hasil ujian
			staffOnly.GET("/exam-reports", resultHandler.ExamReport)

			// ---------- SISWA (CANDIDATE) ----------
			candidate := protected.Group("/candidate", middleware.RequireRoles("student"))
			{
				candidate.GET("/exams", candidateExamHandler.ListExams)

				candidate.POST("/exams/:exam_id/validate-token", candidateExamHandler.ValidateToken)
				candidate.POST("/exams/:exam_id/start", candidateExamHandler.Start)
				candidate.GET("/attempts/:id/questions", attemptEngineHandler.GetQuestions)
				candidate.POST("/attempts/:id/answers", attemptEngineHandler.SaveAnswer)
				candidate.POST("/attempts/:id/questions/:question_id/flag", attemptEngineHandler.Flag)
				candidate.GET("/attempts/:id/discussion", candidateExamHandler.GetDiscussion)
				candidate.POST("/attempts/:id/questions/:question_id/report", candidateExamHandler.ReportQuestion)
				candidate.DELETE("/attempts/:id/questions/:question_id/flag", attemptEngineHandler.Unflag)
				candidate.POST("/attempts/:id/heartbeat", attemptEngineHandler.Heartbeat)
				candidate.POST("/attempts/:id/autosave", attemptEngineHandler.Autosave)
				candidate.POST("/attempts/:id/submit", attemptEngineHandler.Submit)
				candidate.GET("/results", resultHandler.CandidateResults)
			}
			protected.GET("/students/:id/results", resultHandler.StudentResults)

			dash := protected.Group("/dashboard")
			{
				dash.GET("/admin", middleware.RequireRoles("admin"), dashboardHandler.Admin)
				dash.GET("/teacher", middleware.RequireRoles("teacher"), dashboardHandler.Teacher)
				dash.GET("/student", middleware.RequireRoles("student"), dashboardHandler.Student)
			}
		}
	}

	return r, nil
}
