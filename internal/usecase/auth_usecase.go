package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/asepzainudin14/mcbt/internal/model"
	jwtmanager "github.com/asepzainudin14/mcbt/internal/pkg/jwt"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type AuthTokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthResult struct {
	User  *model.User
	Token AuthTokenPair
}

type AuthUsecase struct {
	users  *repository.UserRepository
	tokens *jwtmanager.Manager
}

func NewAuthUsecase(users *repository.UserRepository, tokens *jwtmanager.Manager) *AuthUsecase {
	return &AuthUsecase{users: users, tokens: tokens}
}

func (u *AuthUsecase) Login(ctx context.Context, identifier, password string) (*AuthResult, error) {
	user, err := u.users.FindByIdentifier(ctx, identifier)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user == nil || bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(password),
	) != nil {
		return nil, apperror.New(401, "Invalid credentials", nil)
	}
	if !user.IsActive {
		return nil, apperror.New(403, "Account is disabled", errors.New("inactive user login attempt"))
	}

	pair, err := u.newTokenPair(user)
	if err != nil {
		return nil, err
	}

	if err := u.users.TouchLastLogin(ctx, user.ID); err != nil {
		return nil, apperror.Internal(err)
	}

	return &AuthResult{User: user, Token: pair}, nil
}

func (u *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	claims, err := u.tokens.Parse(refreshToken, jwtmanager.TokenTypeRefresh)
	if err != nil {
		return nil, apperror.New(401, "Invalid refresh token", err)
	}

	user, err := u.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user == nil {
		return nil, apperror.New(401, "Invalid refresh token", errors.New("user no longer exists"))
	}
	if !user.IsActive {
		return nil, apperror.New(403, "Account is disabled", errors.New("inactive user refresh"))
	}
	if user.TokenVersion != claims.TokenVersion {
		return nil, apperror.New(401, "Session revoked, please login again", errors.New("stale token version"))
	}

	pair, err := u.newTokenPair(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Token: pair}, nil
}

func (u *AuthUsecase) Me(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user == nil {
		return nil, apperror.NotFound("User not found", nil)
	}
	return user, nil
}

func (u *AuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		return apperror.Internal(err)
	}
	if user == nil {
		return apperror.NotFound("User not found", nil)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)) != nil {
		return apperror.BadRequest("Old password is incorrect", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal(err)
	}

	if err := u.users.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return apperror.Internal(err)
	}

	return nil
}

func (u *AuthUsecase) newTokenPair(user *model.User) (AuthTokenPair, error) {
	access, err := u.tokens.NewAccessToken(user.ID, user.TokenVersion)
	if err != nil {
		return AuthTokenPair{}, apperror.Internal(err)
	}
	refresh, err := u.tokens.NewRefreshToken(user.ID, user.TokenVersion)
	if err != nil {
		return AuthTokenPair{}, apperror.Internal(err)
	}
	return AuthTokenPair{AccessToken: access, RefreshToken: refresh}, nil
}
