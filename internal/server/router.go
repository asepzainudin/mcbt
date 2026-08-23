package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/config"
	"github.com/asepzainudin14/mcbt/internal/delivery/http/handler"
	jwtmanager "github.com/asepzainudin14/mcbt/internal/pkg/jwt"
	"github.com/asepzainudin14/mcbt/internal/repository"
	middleware "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type RouterDeps struct {
	Cfg *config.Config
	Log *slog.Logger
	DB  *gorm.DB
}

func NewRouter(d RouterDeps) *gin.Engine {
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
		}
	}

	return r
}
