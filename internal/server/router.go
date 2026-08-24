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

	ayRepo := repository.NewAcademicYearRepository(d.DB)
	classRepo := repository.NewClassRepository(d.DB)
	subjectRepo := repository.NewSubjectRepository(d.DB)

	cookies := handler.NewCookieManager(d.Cfg)
	authHandler := handler.NewAuthHandler(authUC, cookies)
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
	mediaRepo := repository.NewMediaRepository(d.DB)

	storageClient, err := storage.NewClient(context.Background(), d.Cfg)
	if err != nil {
		return nil, err
	}

	bankHandler := handler.NewQuestionBankHandler(
		usecase.NewQuestionBankUsecase(bankRepo, subjectRepo, questionRepo),
	)
	questionHandler := handler.NewQuestionHandler(
		usecase.NewQuestionUsecase(questionRepo, bankRepo),
	)
	examRepo := repository.NewExamRepository(d.DB)
	examHandler := handler.NewExamHandler(
		usecase.NewExamUsecase(examRepo, subjectRepo, ayRepo, bankRepo),
	)
	examSectionHandler := handler.NewExamSectionHandler(
		usecase.NewExamSectionUsecase(
			repository.NewExamSectionRepository(d.DB), examRepo, bankRepo,
		),
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
	attemptRepo := repository.NewExamAttemptRepository(d.DB)
	candidateExamHandler := handler.NewCandidateExamHandler(
		usecase.NewCandidateExamUsecase(attemptRepo, studentRepo, examScheduleRepo, examRepo, examParticipantRepo),
	)
	questionImportHandler := handler.NewQuestionImportHandler(
		usecase.NewQuestionImportUsecase(usecase.NewImportTokenStore(), questionRepo),
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
		protected.PUT("/auth/change-password", authHandler.ChangePassword)

		adminOnly := protected.Group("", middleware.RequireRoles("admin"))
		{
			adminOnly.GET("/roles", roleHandler.List)
			adminOnly.POST("/roles/assign", roleHandler.Assign)

			adminOnly.GET("/academic-years", ayHandler.List)
			adminOnly.POST("/academic-years", ayHandler.Create)
			adminOnly.PUT("/academic-years/:id", ayHandler.Update)
			adminOnly.PATCH("/academic-years/:id/activate", ayHandler.Activate)
			adminOnly.DELETE("/academic-years/:id", ayHandler.Delete)

			adminOnly.GET("/classes", classHandler.List)
			adminOnly.POST("/classes", classHandler.Create)
			adminOnly.PUT("/classes/:id", classHandler.Update)
			adminOnly.DELETE("/classes/:id", classHandler.Delete)

			adminOnly.GET("/subjects", subjectHandler.List)
			adminOnly.POST("/subjects", subjectHandler.Create)
			adminOnly.PUT("/subjects/:id", subjectHandler.Update)
			adminOnly.DELETE("/subjects/:id", subjectHandler.Delete)

			adminOnly.GET("/teachers", teacherHandler.List)
			adminOnly.GET("/teachers/import/template", teacherHandler.Template)
			adminOnly.POST("/teachers/import", teacherHandler.Import)
			adminOnly.GET("/teachers/:id", teacherHandler.Get)
			adminOnly.POST("/teachers", teacherHandler.Create)
			adminOnly.PUT("/teachers/:id", teacherHandler.Update)
			adminOnly.DELETE("/teachers/:id", teacherHandler.Delete)

			adminOnly.GET("/students", studentHandler.List)
			adminOnly.GET("/students/import/template", studentHandler.Template)
			adminOnly.POST("/students/import", studentHandler.Import)
			adminOnly.POST("/students/:id/change-class", studentHandler.ChangeClass)
			adminOnly.POST("/students/:id/reset-password", studentHandler.ResetPassword)
			adminOnly.GET("/students/:id", studentHandler.Get)
			adminOnly.POST("/students", studentHandler.Create)
			adminOnly.PUT("/students/:id", studentHandler.Update)
			adminOnly.DELETE("/students/:id", studentHandler.Delete)

			adminOnly.GET("/question-banks", bankHandler.List)
			adminOnly.POST("/question-banks", bankHandler.Create)
			adminOnly.POST("/question-banks/:id/clone", bankHandler.Clone)
			adminOnly.PATCH("/question-banks/:id/publish", bankHandler.Publish)
			adminOnly.PATCH("/question-banks/:id/archive", bankHandler.Archive)
			adminOnly.PUT("/question-banks/:id", bankHandler.Update)
			adminOnly.DELETE("/question-banks/:id", bankHandler.Delete)

			adminOnly.GET("/questions/import/template", questionImportHandler.Template)
			adminOnly.POST("/questions/import/validate", questionImportHandler.Validate)
			adminOnly.POST("/questions/import/process", questionImportHandler.Process)

			adminOnly.POST("/media/upload", mediaHandler.Upload)
			protected.GET("/media/:id/file", mediaHandler.File)

			adminOnly.GET("/questions", questionHandler.List)
			adminOnly.GET("/questions/:id/preview", questionHandler.Preview)
			adminOnly.GET("/questions/:id", questionHandler.Get)
			adminOnly.POST("/question-banks/:id/questions", questionHandler.CreateInBank)
			adminOnly.PUT("/questions/:id", questionHandler.Update)
			adminOnly.DELETE("/questions/:id", questionHandler.Delete)
			adminOnly.PUT("/questions/:id/options/reorder", questionHandler.ReorderOptions)
			adminOnly.PUT("/questions/:id/options/:option_id", questionHandler.UpdateOption)

			adminOnly.GET("/exams", examHandler.List)
			adminOnly.POST("/exams", examHandler.Create)
			adminOnly.GET("/exams/:id", examHandler.Get)
			adminOnly.PUT("/exams/:id", examHandler.Update)
			adminOnly.PUT("/exams/:id/settings", examHandler.UpdateSettings)
			adminOnly.PATCH("/exams/:id/publish", examHandler.Publish)
			adminOnly.PATCH("/exams/:id/close", examHandler.Close)
			adminOnly.DELETE("/exams/:id", examHandler.Delete)

			adminOnly.GET("/exams/:id/sections", examSectionHandler.ListByExam)
			adminOnly.POST("/exams/:id/sections", examSectionHandler.Create)

			adminOnly.PUT("/sections/:id", examSectionHandler.Update)
			adminOnly.DELETE("/sections/:id", examSectionHandler.Delete)
			adminOnly.GET("/sections/:id/questions", examSectionHandler.ListQuestions)
			adminOnly.POST("/sections/:id/questions", examSectionHandler.MapQuestions)

			adminOnly.POST("/exams/:id/schedules", examScheduleHandler.Create)
			adminOnly.GET("/exams/:id/schedules", examScheduleHandler.GetByExam)
			adminOnly.PUT("/schedules/:id", examScheduleHandler.Update)
			adminOnly.DELETE("/schedules/:id", examScheduleHandler.Delete)
			adminOnly.POST("/schedules/:id/generate-token", examScheduleHandler.GenerateToken)

			adminOnly.GET("/exams/:id/participants", examParticipantHandler.List)
			candidate := protected.Group("/candidate", middleware.RequireRoles("student"))
			{
				candidate.GET("/exams", candidateExamHandler.ListExams)
				candidate.POST("/exams/:exam_id/validate-token", candidateExamHandler.ValidateToken)
				candidate.POST("/exams/:exam_id/start", candidateExamHandler.Start)
			}

			adminOnly.POST("/exams/:id/participants/assign-class", examParticipantHandler.AssignClass)
			adminOnly.POST("/exams/:id/participants/assign-individual", examParticipantHandler.AssignIndividual)
			adminOnly.DELETE("/exams/:id/participants/:participant_id", examParticipantHandler.Remove)
			adminOnly.DELETE("/sections/:id/questions/:question_id", examSectionHandler.RemoveQuestion)
		}
	}

	return r, nil
}
