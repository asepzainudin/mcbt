package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

const examTokenCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateExamToken(length int) (string, error) {
	out := make([]byte, length)
	max := big.NewInt(int64(len(examTokenCharset)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = examTokenCharset[n.Int64()]
	}
	return string(out), nil
}

type ExamScheduleUsecase struct {
	schedules ExamScheduleRepo
	exams     ExamRepo
}

func NewExamScheduleUsecase(
	schedules ExamScheduleRepo,
	exams ExamRepo,
) *ExamScheduleUsecase {
	return &ExamScheduleUsecase{schedules: schedules, exams: exams}
}

type ExamScheduleInput struct {
	StartTime time.Time
	EndTime   time.Time
	Token     string
}

func (u *ExamScheduleUsecase) validateTimes(start, end time.Time) error {
	if !end.After(start) {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"end_time": "waktu selesai harus setelah waktu mulai"},
		}
	}
	return nil
}

func (u *ExamScheduleUsecase) examExists(ctx context.Context, examID uuid.UUID) error {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return nil
}

func (u *ExamScheduleUsecase) resolveToken(ctx context.Context, requested string, excludeScheduleID *uuid.UUID) (string, error) {
	token := strings.ToUpper(strings.TrimSpace(requested))
	if token == "" {
		token, _ = generateExamToken(6)
	} else if len(token) < 4 || len(token) > 10 {
		return "", &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"token": "token harus 4–10 karakter"},
		}
	}

	for attempt := 0; attempt < 5; attempt++ {
		exists, err := u.schedules.TokenExists(ctx, token, excludeScheduleID)
		if err != nil {
			return "", apperror.Internal(err)
		}
		if !exists {
			return token, nil
		}
		if strings.TrimSpace(requested) != "" {
			// token eksplisit bentrok -> 409
			return "", apperror.New(409, "Token sudah digunakan jadwal lain", nil)
		}
		token, _ = generateExamToken(6)
	}
	return "", apperror.Internal(fmt.Errorf("gagal membuat token unik"))
}

func (u *ExamScheduleUsecase) Create(ctx context.Context, examID uuid.UUID, in ExamScheduleInput) (*model.ExamSchedule, error) {
	if err := u.examExists(ctx, examID); err != nil {
		return nil, err
	}
	if err := u.validateTimes(in.StartTime, in.EndTime); err != nil {
		return nil, err
	}

	token, err := u.resolveToken(ctx, in.Token, nil)
	if err != nil {
		return nil, err
	}

	schedule := &model.ExamSchedule{
		ExamID:    examID,
		StartTime: in.StartTime,
		EndTime:   in.EndTime,
		Token:     token,
	}
	if err := u.schedules.Create(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (u *ExamScheduleUsecase) Update(ctx context.Context, id uuid.UUID, in ExamScheduleInput) (*model.ExamSchedule, error) {
	schedule, err := u.schedules.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Jadwal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	if err := u.validateTimes(in.StartTime, in.EndTime); err != nil {
		return nil, err
	}

	token, err := u.resolveToken(ctx, in.Token, &id)
	if err != nil {
		return nil, err
	}

	schedule.StartTime = in.StartTime
	schedule.EndTime = in.EndTime
	schedule.Token = token
	if err := u.schedules.Update(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (u *ExamScheduleUsecase) GetByExam(ctx context.Context, examID uuid.UUID) (*model.ExamSchedule, error) {
	if err := u.examExists(ctx, examID); err != nil {
		return nil, err
	}
	schedule, err := u.schedules.FindByExam(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.Internal(err)
	}
	return schedule, nil
}

func (u *ExamScheduleUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.schedules.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Jadwal tidak ditemukan", err)
		}
		return err
	}
	return nil
}

// GenerateToken regenerates a fresh random token for an existing schedule.
func (u *ExamScheduleUsecase) GenerateToken(ctx context.Context, id uuid.UUID) (*model.ExamSchedule, string, error) {
	schedule, err := u.schedules.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", apperror.NotFound("Jadwal tidak ditemukan", err)
		}
		return nil, "", apperror.Internal(err)
	}

	token, err := generateExamToken(6)
	if err != nil {
		return nil, "", apperror.Internal(err)
	}

	schedule.Token = token
	if err := u.schedules.Update(ctx, schedule); err != nil {
		return nil, "", err
	}
	return schedule, token, nil
}
