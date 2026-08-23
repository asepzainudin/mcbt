package jwtmanager

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-secret-key-with-at-least-32-chars!!"

func newTestManager() *Manager {
	return NewManager(testSecret, 15*time.Minute, 7*24*time.Hour)
}

func TestAccessTokenRoundTrip(t *testing.T) {
	m := newTestManager()
	userID := uuid.MustParse("0198f0aa-7b2c-7333-8b4e-1a2b3c4d5e6f")

	token, err := m.NewAccessToken(userID, 3)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	claims, err := m.Parse(token, TokenTypeAccess)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected user id %s, got %s", userID, claims.UserID)
	}
	if claims.TokenVersion != 3 {
		t.Errorf("expected token version 3, got %d", claims.TokenVersion)
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	token, err := m.NewRefreshToken(userID, 1)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	if _, err := m.Parse(token, TokenTypeAccess); err == nil {
		t.Error("access-type parse of refresh token should fail")
	}

	claims, err := m.Parse(token, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("refresh parse failed: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("user id mismatch: %s != %s", claims.UserID, userID)
	}
}

func TestExpiredToken(t *testing.T) {
	m := newTestManager()
	m.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }

	token, err := m.NewAccessToken(uuid.New(), 1)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	m.now = time.Now
	if _, err := m.Parse(token, TokenTypeAccess); err == nil {
		t.Fatal("expected expired token error")
	} else if err != ErrExpiredToken {
		t.Logf("got wrapped error (acceptable): %v", err)
	}
}

func TestTamperedSignature(t *testing.T) {
	m := newTestManager()
	other := NewManager("another-secret-value-that-is-long-enough!", 15*time.Minute, time.Hour)

	token, _ := other.NewAccessToken(uuid.New(), 1)
	if _, err := m.Parse(token, TokenTypeAccess); err == nil {
		t.Fatal("expected signature verification failure")
	}
}
