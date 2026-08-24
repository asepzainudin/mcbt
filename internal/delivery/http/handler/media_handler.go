package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type MediaHandler struct {
	uc *usecase.MediaUsecase
}

func NewMediaHandler(uc *usecase.MediaUsecase) *MediaHandler {
	return &MediaHandler{uc: uc}
}

func (h *MediaHandler) Upload(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}

	uploaderID, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(apperror.BadRequest("File wajib diunggah pada field 'file'", err))
		return
	}

	media, err := h.uc.Upload(c.Request.Context(), fileHeader, c.PostForm("type"), uploaderID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "File berhasil diunggah", media)
}
