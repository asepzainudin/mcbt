package jwtmanager

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	issuer = "mcbt"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

type Claims struct {
	UserID       uuid.UUID
	Type         string
	TokenVersion int
	Registered   jwt.RegisteredClaims
}

type tokenClaims struct {
	jwt.RegisteredClaims
	Type         string `json:"typ"`
	TokenVersion int    `json:"ver"`
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) NewAccessToken(userID uuid.UUID, tokenVersion int) (string, error) {
	return m.sign(userID, TokenTypeAccess, m.accessTTL, tokenVersion)
}

func (m *Manager) NewRefreshToken(userID uuid.UUID, tokenVersion int) (string, error) {
	return m.sign(userID, TokenTypeRefresh, m.refreshTTL, tokenVersion)
}

func (m *Manager) Parse(tokenString, wantType string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(
		tokenString,
		&tokenClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithIssuer(issuer),
		jwt.WithTimeFunc(m.now),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	cl, ok := parsed.Claims.(*tokenClaims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if cl.Type != wantType {
		return nil, fmt.Errorf("%w: wrong token type", ErrInvalidToken)
	}

	userID, err := uuid.Parse(cl.Subject)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid subject", ErrInvalidToken)
	}

	return &Claims{
		UserID:       userID,
		Type:         cl.Type,
		TokenVersion: cl.TokenVersion,
		Registered:   cl.RegisteredClaims,
	}, nil
}

func (m *Manager) sign(userID uuid.UUID, typ string, ttl time.Duration, ver int) (string, error) {
	now := m.now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Type:         typ,
		TokenVersion: ver,
	})
	return token.SignedString(m.secret)
}
