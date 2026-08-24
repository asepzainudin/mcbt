package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
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
	response.Success(c, http.StatusCreated, "File berhasil diunggah", mediaResponse(media))
}

// File streams the stored image through the API (authenticated).
func (h *MediaHandler) File(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	media, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	reader, contentType, size, err := h.uc.GetFile(c.Request.Context(), media.FilePath)
	if err != nil {
		c.Error(apperror.NotFound("File tidak ditemukan", err))
		return
	}
	defer reader.Close()

	c.Header("Cache-Control", "private, max-age=86400")
	c.DataFromReader(http.StatusOK, size, contentType, reader, nil)
}

// mediaResponse builds the API shape: file_path holds the object key,
// url is the authenticated proxy endpoint for rendering <img>.
func mediaResponse(m *model.Media) ginH {
	return ginH{
		"id":          m.ID,
		"file_name":   m.FileName,
		"file_path":   m.FilePath,
		"mime_type":   m.MimeType,
		"file_size":   m.FileSize,
		"url":         "/api/v1/media/" + m.ID.String() + "/file",
		"uploaded_by": m.UploadedBy,
		"created_at":  m.CreatedAt,
	}
}
