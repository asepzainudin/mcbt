package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/asepzainudin14/mcbt/internal/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, http.StatusOK, "System operational", gin.H{
		"status": "UP",
	})
}
