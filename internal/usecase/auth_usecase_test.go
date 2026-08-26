package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	jwtmanager "github.com/asepzainudin14/mcbt/internal/pkg/jwt"
)

func newAuthUC() (*AuthUsecase, *fakeUserRepo) {
	tokens := jwtmanager.NewManager("test-secret-auth", 15*time.Minute, 7*24*time.Hour)
	users := &fakeUserRepo{}
	return NewAuthUsecase(users, tokens), users
}

func hashPwd(t *testing.T, pwd string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return string(h)
}

func activeUser(t *testing.T, password string) *model.User {
	t.Helper()
	return &model.User{
		BaseModel:    model.BaseModel{ID: uuid.New()},
		Username:     "budi",
		Name:         "Budi",
		Email:        "budi@x.id",
		PasswordHash: hashPwd(t, password),
		IsActive:     true,
		TokenVersion: 1,
	}
}

func TestAuthLogin_Success(t *testing.T) {
	uc, users := newAuthUC()
	user := activeUser(t, "Rahasia@123")
	touched := false
	users.findByIdentifierFn = func(ctx context.Context, id string) (*model.User, error) {
		return user, nil
	}
	users.touchLastLoginFn = func(ctx context.Context, id uuid.UUID) error {
		touched = true
		return nil
	}

	res, err := uc.Login(ctxBg(), "budi", "Rahasia@123")
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if res.User.ID != user.ID {
		t.Errorf("user tidak sesuai")
	}
	if res.Token.AccessToken == "" || res.Token.RefreshToken == "" {
		t.Error("token pair kosong")
	}
	if !touched {
		t.Error("TouchLastLogin tidak dipanggil")
	}
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	uc, users := newAuthUC()
	users.findByIdentifierFn = func(context.Context, string) (*model.User, error) {
		return activeUser(t, "Rahasia@123"), nil
	}
	_, err := uc.Login(ctxBg(), "budi", "salah")
	if ae := apperror.From(err); ae.Code != 401 {
		t.Fatalf("want 401 invalid credentials, got %v", err)
	}
}

func TestAuthLogin_UserNotFound(t *testing.T) {
	uc, _ := newAuthUC()
	_, err := uc.Login(ctxBg(), "takada", "x")
	if ae := apperror.From(err); ae.Code != 401 {
		t.Fatalf("want 401, got %v", err)
	}
}

func TestAuthLogin_InactiveAccount(t *testing.T) {
	uc, users := newAuthUC()
	u := activeUser(t, "Rahasia@123")
	u.IsActive = false
	users.findByIdentifierFn = func(context.Context, string) (*model.User, error) { return u, nil }

	_, err := uc.Login(ctxBg(), "budi", "Rahasia@123")
	if ae := apperror.From(err); ae.Code != 403 {
		t.Fatalf("want 403 akun nonaktif, got %v", err)
	}
}

func TestRefresh_RotatesToken(t *testing.T) {
	uc, users := newAuthUC()
	user := activeUser(t, "Rahasia@123")
	users.findByIdentifierFn = func(context.Context, string) (*model.User, error) { return user, nil }
	users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) { return user, nil }

	first, err := uc.Login(ctxBg(), "budi", "Rahasia@123")
	if err != nil {
		t.Fatal(err)
	}
	res, err := uc.Refresh(ctxBg(), first.Token.RefreshToken)
	if err != nil {
		t.Fatalf("refresh error: %v", err)
	}
	if res.Token.AccessToken == "" {
		t.Error("access token baru kosong")
	}
}

func TestRefresh_RejectsAccessTokenAsRefresh(t *testing.T) {
	uc, _ := newAuthUC()
	login, _ := uc.Login(ctxBg(), "nopass-user", "") // gagal → tak dipakai
	_ = login

	mgr := jwtmanager.NewManager("test-secret-auth", time.Minute, time.Minute)
	uc2 := NewAuthUsecase(&fakeUserRepo{}, mgr)
	access, _ := mgr.NewAccessToken(uuid.New(), 1)

	_, err := uc2.Refresh(ctxBg(), access)
	if ae := apperror.From(err); ae.Code != 401 {
		t.Fatalf("access token dipakai sbg refresh harus 401, got %v", err)
	}
}

func TestRefresh_StaleTokenVersion(t *testing.T) {
	uc, users := newAuthUC()
	user := activeUser(t, "Rahasia@123") // TokenVersion 1
	users.findByIdentifierFn = func(context.Context, string) (*model.User, error) { return user, nil }
	users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) {
		stale := *user
		stale.TokenVersion = 99 // password direset → versi naik
		return &stale, nil
	}

	login, err := uc.Login(ctxBg(), "budi", "Rahasia@123")
	if err != nil {
		t.Fatal(err)
	}
	_, err = uc.Refresh(ctxBg(), login.Token.RefreshToken)
	if ae := apperror.From(err); ae.Code != 401 {
		t.Fatalf("token lama (versi beda) harus 401, got %v", err)
	}
}

func TestMe_FoundAndNotFound(t *testing.T) {
	uc, users := newAuthUC()
	id := uuid.New()
	users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	if _, err := uc.Me(ctxBg(), id); err == nil {
		t.Error("user tak ada harus error")
	}

	users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) {
		return activeUser(t, "x"), nil
	}
	if _, err := uc.Me(ctxBg(), id); err != nil {
		t.Errorf("me error: %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	t.Run("sukses memperbarui hash", func(t *testing.T) {
		uc, users := newAuthUC()
		var savedHash string
		user := activeUser(t, "Lama@123")
		users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) { return user, nil }
		users.updatePasswordFn = func(_ context.Context, _ uuid.UUID, h string) error {
			savedHash = h
			return nil
		}

		if err := uc.ChangePassword(ctxBg(), user.ID, "Lama@123", "Baru@12345"); err != nil {
			t.Fatalf("change password: %v", err)
		}
		if bcrypt.CompareHashAndPassword([]byte(savedHash), []byte("Baru@12345")) != nil {
			t.Error("hash tersimpan tidak cocok dengan password baru")
		}
	})

	t.Run("password lama salah → 400", func(t *testing.T) {
		uc, users := newAuthUC()
		user := activeUser(t, "Lama@123")
		users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) { return user, nil }

		err := uc.ChangePassword(ctxBg(), user.ID, "Salah@123", "Baru@12345")
		if ae := apperror.From(err); ae.Code != 400 {
			t.Fatalf("want 400, got %v", err)
		}
	})

	t.Run("user tidak ada → 404", func(t *testing.T) {
		uc, users := newAuthUC()
		// kontrak repo: tak ditemukan → (nil, nil)
		users.findByIDFn = func(context.Context, uuid.UUID) (*model.User, error) { return nil, nil }

		err := uc.ChangePassword(ctxBg(), uuid.New(), "a", "b")
		if ae := apperror.From(err); ae.Code != 404 {
			t.Fatalf("want 404, got %v", err)
		}
	})
}
