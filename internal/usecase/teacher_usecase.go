package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	passwordutil "github.com/asepzainudin14/mcbt/internal/pkg/password"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

var teacherTemplateColumns = []string{"username", "name", "email", "nip", "phone"}

type TeacherUsecase struct {
	repo  *repository.TeacherRepository
	roles *repository.RoleRepository
}

func NewTeacherUsecase(repo *repository.TeacherRepository, roles *repository.RoleRepository) *TeacherUsecase {
	return &TeacherUsecase{repo: repo, roles: roles}
}

func (u *TeacherUsecase) List(ctx context.Context, search string, page, limit int) ([]model.Teacher, int64, error) {
	result, err := u.repo.List(ctx, search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *TeacherUsecase) Get(ctx context.Context, id uuid.UUID) (*model.Teacher, error) {
	t, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Guru tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	return t, nil
}

func (u *TeacherUsecase) hashDefault() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordutil.DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (u *TeacherUsecase) Create(ctx context.Context, in repository.TeacherUpsert) (*model.Teacher, error) {
	field, dup, err := u.repo.ExistsDuplicate(ctx, in.Username, in.Email, deref(in.Nip), nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, field+" sudah digunakan", nil)
	}

	role, err := u.roles.FindByName(ctx, "teacher")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(422, "Role teacher belum tersedia", err)
		}
		return nil, apperror.Internal(err)
	}

	in.PasswordHash, err = u.hashDefault()
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return u.repo.CreateWithUser(ctx, in, role.ID)
}

func (u *TeacherUsecase) Update(ctx context.Context, id uuid.UUID, in repository.TeacherUpdate) (*model.Teacher, error) {
	teacher, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Guru tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	user := teacher.User
	field, dup, err := u.repo.ExistsDuplicate(
		ctx,
		strOrDefault(user.Username),
		in.Email,
		deref(in.Nip),
		&teacher.UserID,
	)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, field+" sudah digunakan", nil)
	}

	if err := u.repo.Update(ctx, teacher, in); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *TeacherUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	teacher, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Guru tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return u.repo.DeleteWithUser(ctx, teacher)
}

type ImportRowError struct {
	Row    int    `json:"row"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ImportResult struct {
	ImportedCount int              `json:"imported_count"`
	Skipped       []ImportRowError `json:"skipped"`
}

func (u *TeacherUsecase) Import(ctx context.Context, fileBytes []byte) (*ImportResult, error) {
	rows, err := parseSheet(fileBytes, teacherTemplateColumns)
	if err != nil {
		return nil, apperror.BadRequest(err.Error(), nil)
	}

	var (
		valid   []repository.TeacherUpsert
		skipped []ImportRowError
	)

	for i, row := range rows {
		rowNo := i + 2
		if err := validateRequired(map[string]string{
			"username": row[0], "name": row[1], "email": row[2],
		}); err != nil {
			skipped = append(skipped, ImportRowError{Row: rowNo, Field: err.Error(), Reason: "wajib diisi"})
			continue
		}
		field, dup, err := u.repo.ExistsDuplicate(ctx, row[0], row[2], row[3], nil)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if dup {
			skipped = append(skipped, ImportRowError{Row: rowNo, Field: field, Reason: "sudah digunakan"})
			continue
		}

		hash, err := u.hashDefault()
		if err != nil {
			return nil, apperror.Internal(err)
		}
		valid = append(valid, repository.TeacherUpsert{
			Username:     row[0],
			Name:         row[1],
			Email:        row[2],
			Nip:          strPtr(row[3]),
			Phone:        strPtr(row[4]),
			PasswordHash: hash,
		})
	}

	role, err := u.roles.FindByName(ctx, "teacher")
	if err != nil {
		return nil, apperror.New(422, "Role teacher belum tersedia", err)
	}

	if err := u.repo.CreateManyWithUsers(ctx, valid, role.ID); err != nil {
		return nil, err
	}

	return &ImportResult{ImportedCount: len(valid), Skipped: skipped}, nil
}

func (u *TeacherUsecase) TemplateXLSX() ([]byte, error) {
	return buildTemplate("data_guru", teacherTemplateColumns, [][]any{
		{"guru001", "Budi Santoso", "budi@sekolah.id", "197001012000031001", "081234567890"},
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strOrDefault(s string) string {
	return s
}

func validateRequired(fields map[string]string) error {
	for name, val := range fields {
		if val == "" {
			return fmt.Errorf("%s", name)
		}
	}
	return nil
}

func parseSheet(fileBytes []byte, columns []string) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("file Excel tidak dapat dibaca")
	}
	defer f.Close()

	sheet := f.GetSheetList()[0]
	tableRows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("sheet kosong atau tidak terbaca")
	}
	if len(tableRows) < 1 {
		return nil, fmt.Errorf("file harus memiliki baris header")
	}

	header := normalizeHeader(tableRows[0])
	want := make([]string, len(columns))
	for i, c := range columns {
		want[i] = c
	}
	if !equalFoldSlice(header[:min(len(header), len(want))], want[:min(len(header), len(want))]) {
		return nil, fmt.Errorf("header harus berisi kolom: %s", joinColumns(columns))
	}

	var data [][]string
	for _, r := range tableRows[1:] {
		row := make([]string, len(columns))
		for i := range columns {
			if i < len(r) {
				row[i] = trimSpace(r[i])
			}
		}
		data = append(data, row)
	}
	return data, nil
}

func normalizeHeader(cells []string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = lower(trimSpace(c))
	}
	return out
}

func equalFoldSlice(a, b []string) bool {
	for i := range b {
		if i >= len(a) || a[i] != b[i] {
			return false
		}
	}
	return true
}

func joinColumns(cols []string) string {
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += ", "
		}
		out += c
	}
	return out
}

func buildTemplate(sheetName string, columns []string, sample [][]any) ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if sheetName != "Sheet1" {
		f.SetSheetName(sheet, sheetName)
		sheet = sheetName
	}
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, col)
	}
	for rIdx, row := range sample {
		for cIdx, v := range row {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	widths := map[int]float64{1: 16, 2: 24, 3: 28, 4: 24, 5: 16}
	for idx, w := range widths {
		col, _ := excelize.ColumnNumberToName(idx + 1)
		_ = f.SetColWidth(sheet, col, col, w)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func trimSpace(s string) string { return strings.TrimSpace(s) }

func lower(s string) string { return strings.ToLower(s) }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
