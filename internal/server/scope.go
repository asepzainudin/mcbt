package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

// dataScopeGuard memastikan guru hanya mengakses resource miliknya.
// Admin lolos tanpa pemeriksaan. kind: bank|exam|section|attempt|question
func dataScopeGuard(access *usecase.AccessUsecase, kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := mw.CurrentPrincipal(c)
		if !ok {
			c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
			c.Abort()
			return
		}
		if principal.HasRole("admin") {
			c.Next()
			return
		}

		userID, err := uuid.Parse(principal.UserID)
		if err != nil {
			c.Error(apperror.BadRequest("Invalid user id", err))
			c.Abort()
			return
		}
		resID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.Error(apperror.BadRequest("ID tidak valid", err))
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		var assertErr error
		switch kind {
		case "bank":
			assertErr = access.AssertBankOwner(ctx, userID, false, resID)
		case "exam":
			assertErr = access.AssertExamOwner(ctx, userID, false, resID)
		case "section":
			assertErr = access.AssertSectionOwner(ctx, userID, false, resID)
		case "attempt":
			assertErr = access.AssertAttemptOwner(ctx, userID, false, resID)
		case "question":
			assertErr = access.AssertQuestionOwner(ctx, userID, false, resID)
		default:
			assertErr = apperror.Forbidden("Akses ditolak", nil)
		}
		if assertErr != nil {
			c.Error(assertErr)
			c.Abort()
			return
		}
		c.Next()
	}
}
