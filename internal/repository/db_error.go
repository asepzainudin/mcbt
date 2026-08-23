package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func TranslateDBError(err error, conflictMessage string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("Data tidak ditemukan", err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return apperror.New(409, conflictMessage, err)
		case pgForeignKeyViolation:
			return apperror.New(409, "Data sedang digunakan oleh data lain dan tidak dapat dihapus", err)
		}
	}

	return apperror.Internal(err)
}
