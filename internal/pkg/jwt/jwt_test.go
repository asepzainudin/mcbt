package jwtmanager

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestManager() *Manager {
	return NewManager("unit-test-secret", 15*time.Minute, 7*24*time.Hour)
}

func TestAccessTokenRoundtrip(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	token, err := m.NewAccessToken(userID, 3)
	if err != nil {
		t.Fatalf("NewAccessToken error: %v", err)
	}

	claims, err := m.Parse(token, "access")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if claims.Registered.Subject != userID.String() {
		t.Errorf("subject = %q, want %q", claims.Registered.Subject, userID)
	}
	if claims.TokenVersion != 3 {
		t.Errorf("token_version = %d, want 3", claims.TokenVersion)
	}
}

func TestRefreshTokenRoundtrip(t *testing.T) {
	m := newTestManager()
	userID := uuid.New()

	token, err := m.NewRefreshToken(userID, 1)
	if err != nil {
		t.Fatalf("NewRefreshToken error: %v", err)
	}
	if _, err := m.Parse(token, "refresh"); err != nil {
		t.Fatalf("Parse refresh error: %v", err)
	}
}

func TestParse_RejectsWrongType(t *testing.T) {
	m := newTestManager()
	token, _ := m.NewAccessToken(uuid.New(), 0)

	if _, err := m.Parse(token, "refresh"); err == nil {
		t.Error("access token diterima sebagai refresh — seharusnya ditolak")
	}
}

func TestParse_RejectsGarbage(t *testing.T) {
	m := newTestManager()
	if _, err := m.Parse("bukan.token.jwt", "access"); err == nil {
		t.Error("token sampah diterima — seharusnya error")
	}
}

func TestParse_RejectsWrongSecret(t *testing.T) {
	a := NewManager("secret-a", time.Minute, time.Minute)
	b := NewManager("secret-b", time.Minute, time.Minute)

	token, _ := a.NewAccessToken(uuid.New(), 0)
	if _, err := b.Parse(token, "access"); err == nil {
		t.Error("token dari secret berbeda diterima — seharusnya ditolak")
	}
}
