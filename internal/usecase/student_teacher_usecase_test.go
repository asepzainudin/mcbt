package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

func newStudentUC() (*StudentUsecase, *fakeStudentRepo, *fakeRoleRepo, *fakeClassRepo) {
	repo := &fakeStudentRepo{}
	roles := &fakeRoleRepo{}
	classes := &fakeClassRepo{}
	return NewStudentUsecase(repo, roles, classes), repo, roles, classes
}

func studentWithUser(name string) *model.Student {
	return &model.Student{
		BaseModel: model.BaseModel{ID: uuid.New()},
		UserID:    uuid.New(),
		Nis:       "20250001",
		User:      &model.User{BaseModel: model.BaseModel{ID: uuid.New()}, Username: strings.ToLower(name), Name: name, Email: name + "@x.id"},
	}
}

func TestStudentGet_NotFound(t *testing.T) {
	uc, repo, _, _ := newStudentUC()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) { return nil, notFound() }

	_, err := uc.Get(ctxBg(), uuid.New())
	if ae := apperror.From(err); ae.Code != 404 {
		t.Fatalf("want 404, got %v", err)
	}
}

func TestStudentCreate_ClassValidation(t *testing.T) {
	uc, _, _, classes := newStudentUC()
	classes.findByIDFn = func(context.Context, uuid.UUID) (*model.Class, error) { return nil, notFound() }

	classID := uuid.New()
	_, err := uc.Create(ctxBg(), repository.StudentUpsert{
		Username: "baru", Name: "Baru", Email: "baru@x.id", Nis: "20259999", PasswordHash: "x", ClassID: &classID,
	})
	ae := apperror.From(err)
	if ae.Code != 422 || ae.Details["class_id"] != "kelas tidak ditemukan" {
		t.Fatalf("want 422 class_id, got %v", err)
	}
}

func TestStudentCreate_Duplicate(t *testing.T) {
	uc, repo, _, _ := newStudentUC()
	repo.existsDupFn = func(context.Context, string, string, string, *uuid.UUID) (string, bool, error) {
		return "email", true, nil
	}

	_, err := uc.Create(ctxBg(), repository.StudentUpsert{
		Username: "dup", Name: "Dup", Email: "dup@x.id", Nis: "20250002", PasswordHash: "x",
	})
	if ae := apperror.From(err); ae.Code != 409 {
		t.Fatalf("want 409, got %v", err)
	}
}

func TestStudentResetPassword(t *testing.T) {
	t.Run("custom password dipakai", func(t *testing.T) {
		uc, repo, _, _ := newStudentUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) {
			return studentWithUser("Satu"), nil
		}

		res, err := uc.ResetPassword(ctxBg(), uuid.New(), "Custom@123")
		if err != nil {
			t.Fatalf("reset: %v", err)
		}
		if res.NewPassword != "Custom@123" {
			t.Errorf("want custom password, got %q", res.NewPassword)
		}
		if len(repo.resetted) != 1 {
			t.Error("UpdatePasswordByUser tidak dipanggil")
		}
	})

	t.Run("kosong → generate 10 karakter", func(t *testing.T) {
		uc, repo, _, _ := newStudentUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) {
			return studentWithUser("Dua"), nil
		}

		res, err := uc.ResetPassword(ctxBg(), uuid.New(), "  ")
		if err != nil {
			t.Fatalf("reset: %v", err)
		}
		if len(res.NewPassword) != 10 {
			t.Errorf("panjang generated = %d, want 10", len(res.NewPassword))
		}
	})

	t.Run("custom <8 karakter → 422", func(t *testing.T) {
		uc, repo, _, _ := newStudentUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) {
			return studentWithUser("Tiga"), nil
		}

		_, err := uc.ResetPassword(ctxBg(), uuid.New(), "pendek")
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("siswa tidak ada → 404", func(t *testing.T) {
		uc, repo, _, _ := newStudentUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) { return nil, notFound() }

		_, err := uc.ResetPassword(ctxBg(), uuid.New(), "")
		if ae := apperror.From(err); ae.Code != 404 {
			t.Fatalf("want 404, got %v", err)
		}
	})
}

func TestStudentChangeClass(t *testing.T) {
	t.Run("kelas tujuan tak ada → 422", func(t *testing.T) {
		uc, _, _, classes := newStudentUC()
		classes.findByIDFn = func(context.Context, uuid.UUID) (*model.Class, error) { return nil, notFound() }

		_, err := uc.ChangeClass(ctxBg(), uuid.New(), uuid.New())
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("siswa tak ada → 404", func(t *testing.T) {
		uc, repo, _, classes := newStudentUC()
		classes.findByIDFn = func(context.Context, uuid.UUID) (*model.Class, error) { return &model.Class{}, nil }
		repo.changeClassFn = func(context.Context, uuid.UUID, uuid.UUID) (*model.Student, error) {
			return nil, notFound()
		}

		_, err := uc.ChangeClass(ctxBg(), uuid.New(), uuid.New())
		if ae := apperror.From(err); ae.Code != 404 {
			t.Fatalf("want 404, got %v", err)
		}
	})
}

func TestStudentDelete(t *testing.T) {
	uc, repo, _, _ := newStudentUC()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Student, error) {
		return studentWithUser("Hapus"), nil
	}
	deleted := false
	repo.deleteFn = func(context.Context, *model.Student) error { deleted = true; return nil }

	if err := uc.Delete(ctxBg(), uuid.New()); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleted {
		t.Error("DeleteWithUser tidak dipanggil")
	}
}

// ---------- Teacher ----------

func newTeacherUC() (*TeacherUsecase, *fakeTeacherRepo, *fakeRoleRepo) {
	repo := &fakeTeacherRepo{}
	roles := &fakeRoleRepo{}
	return NewTeacherUsecase(repo, roles), repo, roles
}

func teacherWithUser(name string) *model.Teacher {
	return &model.Teacher{
		BaseModel: model.BaseModel{ID: uuid.New()},
		UserID:    uuid.New(),
		User:      &model.User{BaseModel: model.BaseModel{ID: uuid.New()}, Username: strings.ToLower(name), Name: name, Email: name + "@x.id"},
	}
}

func TestTeacherResetPassword(t *testing.T) {
	t.Run("custom password", func(t *testing.T) {
		uc, repo, _ := newTeacherUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Teacher, error) {
			return teacherWithUser("Guru"), nil
		}

		res, err := uc.ResetPassword(ctxBg(), uuid.New(), "GuruBaru@1")
		if err != nil {
			t.Fatalf("reset: %v", err)
		}
		if res.NewPassword != "GuruBaru@1" {
			t.Errorf("want custom, got %q", res.NewPassword)
		}
		if len(repo.resetted) != 1 {
			t.Error("password tidak diupdate")
		}
	})

	t.Run("pendek → 422", func(t *testing.T) {
		uc, repo, _ := newTeacherUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Teacher, error) {
			return teacherWithUser("Guru"), nil
		}

		_, err := uc.ResetPassword(ctxBg(), uuid.New(), "abc")
		if ae := apperror.From(err); ae.Code != 422 {
			t.Fatalf("want 422, got %v", err)
		}
	})

	t.Run("guru tak ada → 404", func(t *testing.T) {
		uc, repo, _ := newTeacherUC()
		repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Teacher, error) { return nil, notFound() }

		_, err := uc.ResetPassword(ctxBg(), uuid.New(), "")
		if ae := apperror.From(err); ae.Code != 404 {
			t.Fatalf("want 404, got %v", err)
		}
	})
}

func TestTeacherGet_NotFound(t *testing.T) {
	uc, repo, _ := newTeacherUC()
	repo.findByIDFn = func(context.Context, uuid.UUID) (*model.Teacher, error) { return nil, notFound() }

	_, err := uc.Get(ctxBg(), uuid.New())
	if ae := apperror.From(err); ae.Code != 404 {
		t.Fatalf("want 404, got %v", err)
	}
}

var _ = gorm.ErrRecordNotFound // jaga bila dipakai test tambahan
