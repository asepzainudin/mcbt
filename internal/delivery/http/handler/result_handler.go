package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/repository"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ResultHandler struct {
	uc       *usecase.ResultUsecase
	students *repository.StudentRepository
}

func NewResultHandler(uc *usecase.ResultUsecase, students *repository.StudentRepository) *ResultHandler {
	return &ResultHandler{uc: uc, students: students}
}

// CandidateResults: hasil ujian siswa yang sedang login.
func (h *ResultHandler) CandidateResults(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}
	userID, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return
	}

	student, err := h.students.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.Error(apperror.New(http.StatusForbidden, "Akun ini bukan siswa", nil))
		return
	}

	rows, err := h.uc.StudentResults(c.Request.Context(), student.ID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Hasil ujian", rows)
}

// ExamReport: laporan ujian seluruh siswa dengan filter.
func (h *ResultHandler) ExamReport(c *gin.Context) {
	f := repository.ExamReportFilter{
		Page:  1,
		Limit: 20,
	}
	if raw := c.Query("exam_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("exam_id harus UUID yang valid", err))
			return
		}
		f.ExamID = &id
	}
	if raw := c.Query("subject_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("subject_id harus UUID yang valid", err))
			return
		}
		f.SubjectID = &id
	}
	if raw := c.Query("class_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("class_id harus UUID yang valid", err))
			return
		}
		f.ClassID = &id
	}
	if raw := c.Query("academic_year_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("academic_year_id harus UUID yang valid", err))
			return
		}
		f.AcademicYearID = &id
	}
	if raw := c.Query("date_from"); raw != "" {
		f.DateFrom = &raw
	}
	if raw := c.Query("date_to"); raw != "" {
		f.DateTo = &raw
	}
	if raw := c.Query("page"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			f.Page = v
		}
	}
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			f.Limit = v
		}
	}

	result, err := h.uc.ExamReport(c.Request.Context(), f)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Laporan ujian", result)
}

// ExamResults: rekap nilai + ranking (admin).
func (h *ResultHandler) ExamResults(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	var classID *uuid.UUID
	if raw := c.Query("class_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("class_id harus UUID yang valid", err))
			return
		}
		classID = &id
	}

	ranked, err := h.uc.ExamResults(c.Request.Context(), examID, classID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Rekap nilai", ranked)
}

// StudentResults: hasil ujian siswa (siswa lihat miliknya, admin lihat siapa pun).
func (h *ResultHandler) StudentResults(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}

	requestedID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("student_id tidak valid", err))
		return
	}

	// siswa hanya boleh lihat hasil sendiri
	userID, _ := uuid.Parse(principal.UserID)
	isAdmin := false
	for _, r := range principal.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		student, err := h.students.FindByUserID(c.Request.Context(), userID)
		if err != nil || student == nil || student.ID != requestedID {
			c.Error(apperror.New(http.StatusForbidden, "Anda hanya dapat melihat hasil sendiri", nil))
			return
		}
	}

	rows, err := h.uc.StudentResults(c.Request.Context(), requestedID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Hasil ujian siswa", rows)
}

// PublishResults: toggle hasil ujian.
func (h *ResultHandler) PublishResults(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}

	var req struct {
		Published *bool `json:"published" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if err := h.uc.PublishResults(c.Request.Context(), examID, *req.Published); err != nil {
		c.Error(err)
		return
	}
	msg := "Hasil ujian disembunyikan"
	if *req.Published {
		msg = "Hasil ujian dipublikasikan ke peserta"
	}
	response.Success(c, http.StatusOK, msg, ginH{"results_published": *req.Published})
}
