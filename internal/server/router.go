package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/config"
	"github.com/asepzainudin14/mcbt/internal/delivery/http/handler"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	middleware "github.com/asepzainudin14/mcbt/internal/server/middleware"
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

	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(d.Log))
	r.Use(middleware.Recovery(d.Log))
	r.Use(middleware.ErrorHandler(d.Log))

	v1 := r.Group("/api/v1")
	registerHealthRoutes(v1)

	r.NoRoute(func(c *gin.Context) {
		c.Error(apperror.NotFound("Route not found", nil))
	})

	return r
}

func registerHealthRoutes(rg *gin.RouterGroup) {
	h := handler.NewHealthHandler()
	rg.GET("/health", h.Health)
}
