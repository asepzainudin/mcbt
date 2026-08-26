package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	passwordutil "github.com/asepzainudin14/mcbt/internal/pkg/password"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

var studentTemplateColumns = []string{"username", "name", "email", "nis", "phone", "class_name"}

type StudentUsecase struct {
	repo    *repository.StudentRepository
	roles   *repository.RoleRepository
	classes *repository.ClassRepository
}

func NewStudentUsecase(repo *repository.StudentRepository, roles *repository.RoleRepository, classes *repository.ClassRepository) *StudentUsecase {
	return &StudentUsecase{repo: repo, roles: roles, classes: classes}
}

func (u *StudentUsecase) List(ctx context.Context, search string, classID *uuid.UUID, page, limit int) ([]model.Student, int64, error) {
	result, err := u.repo.List(ctx, search, classID, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *StudentUsecase) Get(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	s, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	return s, nil
}

func (u *StudentUsecase) validateClass(ctx context.Context, classID *uuid.UUID) error {
	if classID == nil {
		return nil
	}
	if _, err := u.classes.FindByID(ctx, *classID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperror.AppError{
				Code:    apperror.CodeUnprocessable,
				Message: "Validasi gagal",
				Details: map[string]string{"class_id": "kelas tidak ditemukan"},
			}
		}
		return apperror.Internal(err)
	}
	return nil
}

func (u *StudentUsecase) hashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (u *StudentUsecase) Create(ctx context.Context, in repository.StudentUpsert) (*model.Student, error) {
	field, dup, err := u.repo.ExistsDuplicate(ctx, in.Username, in.Email, in.Nis, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, field+" sudah digunakan", nil)
	}
	if err := u.validateClass(ctx, in.ClassID); err != nil {
		return nil, err
	}

	in.PasswordHash, err = u.hashPassword(passwordutil.DefaultPassword)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	role, err := u.roles.FindByName(ctx, "student")
	if err != nil {
		return nil, apperror.New(422, "Role student belum tersedia", err)
	}

	return u.repo.CreateWithUser(ctx, in, role.ID)
}

func (u *StudentUsecase) Update(ctx context.Context, id uuid.UUID, in repository.StudentUpdate) (*model.Student, error) {
	student, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	user := student.User
	field, dup, err := u.repo.ExistsDuplicate(
		ctx,
		strOrDefault(user.Username),
		in.Email,
		in.Nis,
		&student.UserID,
	)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, field+" sudah digunakan", nil)
	}
	if err := u.validateClass(ctx, in.ClassID); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, student, in); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *StudentUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	student, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return u.repo.DeleteWithUser(ctx, student)
}

func (u *StudentUsecase) ChangeClass(ctx context.Context, studentID, targetClassID uuid.UUID) (*model.Student, error) {
	if _, err := u.classes.FindByID(ctx, targetClassID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(422, "Kelas tujuan tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	student, err := u.repo.ChangeClass(ctx, studentID, targetClassID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return nil, err
	}
	return student, nil
}

type ResetPasswordResult struct {
	NewPassword string `json:"new_password"`
}

func (u *StudentUsecase) ResetPassword(ctx context.Context, id uuid.UUID, custom string) (*ResetPasswordResult, error) {
	student, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	newPassword := strings.TrimSpace(custom)
	if newPassword == "" {
		newPassword, err = passwordutil.Generate(10)
		if err != nil {
			return nil, apperror.Internal(err)
		}
	} else if len(newPassword) < 8 {
		return nil, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"new_password": "password minimal 8 karakter"},
		}
	}

	hash, err := u.hashPassword(newPassword)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	if err := u.repo.UpdatePasswordByUser(ctx, student.UserID, hash); err != nil {
		return nil, apperror.Internal(err)
	}

	return &ResetPasswordResult{NewPassword: newPassword}, nil
}

func (u *StudentUsecase) Import(ctx context.Context, fileBytes []byte) (*ImportResult, error) {
	rows, err := parseSheet(fileBytes, studentTemplateColumns)
	if err != nil {
		return nil, apperror.BadRequest(err.Error(), nil)
	}

	allClasses, err := u.classes.ListAll(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	classByName := make(map[string]uuid.UUID, len(allClasses))
	for _, c := range allClasses {
		classByName[lower(c.Name)] = c.ID
	}

	var (
		valid   []repository.StudentUpsert
		skipped []ImportRowError
		seen    = map[string]bool{}
	)

	for i, row := range rows {
		rowNo := i + 2

		missing := ""
		for _, idx := range []int{0, 1, 2, 3} {
			if row[idx] == "" {
				missing = studentTemplateColumns[idx]
				break
			}
		}
		if missing != "" {
			skipped = append(skipped, ImportRowError{Row: rowNo, Field: missing, Reason: "wajib diisi"})
			continue
		}

		key := lower(row[0]) + "|" + lower(row[2]) + "|" + row[3]
		if seen[key] {
			skipped = append(skipped, ImportRowError{Row: rowNo, Field: "-", Reason: "duplikat di dalam file"})
			continue
		}
		seen[key] = true

		var classID *uuid.UUID
		if row[5] != "" {
			id, ok := classByName[lower(row[5])]
			if !ok {
				skipped = append(skipped, ImportRowError{Row: rowNo, Field: "class_name", Reason: "kelas tidak ditemukan"})
				continue
			}
			classID = &id
		}

		field, dup, err := u.repo.ExistsDuplicate(ctx, row[0], row[2], row[3], nil)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if dup {
			skipped = append(skipped, ImportRowError{Row: rowNo, Field: field, Reason: "sudah digunakan"})
			continue
		}

		hash, err := u.hashPassword(passwordutil.DefaultPassword)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		valid = append(valid, repository.StudentUpsert{
			Username:     row[0],
			Name:         row[1],
			Email:        row[2],
			Nis:          row[3],
			Phone:        strPtr(row[4]),
			ClassID:      classID,
			PasswordHash: hash,
		})
	}

	role, err := u.roles.FindByName(ctx, "student")
	if err != nil {
		return nil, apperror.New(422, "Role student belum tersedia", err)
	}

	if err := u.repo.CreateManyWithUsers(ctx, valid, role.ID); err != nil {
		return nil, err
	}

	return &ImportResult{ImportedCount: len(valid), Skipped: skipped}, nil
}

func (u *StudentUsecase) TemplateXLSX() ([]byte, error) {
	return buildTemplate("data_siswa", studentTemplateColumns, [][]any{
		{"siswa001", "Ani Lestari", "ani@sekolah.id", "2025001", "081298765432", "XII IPA 1"},
	})
}
