package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebson1988/chirpy/internal/auth"
	"github.com/glebson1988/chirpy/internal/database"
	"github.com/google/uuid"
)

type stubDB struct {
	getUserFromRefreshToken func(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error)
	revokeRefreshToken      func(ctx context.Context, token string) error
}

func (s *stubDB) GetUserFromRefreshToken(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error) {
	if s.getUserFromRefreshToken == nil {
		return database.GetUserFromRefreshTokenRow{}, errors.New("not implemented")
	}
	return s.getUserFromRefreshToken(ctx, token)
}

func (s *stubDB) RevokeRefreshToken(ctx context.Context, token string) error {
	if s.revokeRefreshToken == nil {
		return errors.New("not implemented")
	}
	return s.revokeRefreshToken(ctx, token)
}

func TestHandlerRefreshSuccess(t *testing.T) {
	userID := uuid.New()
	db := &stubDB{
		getUserFromRefreshToken: func(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error) {
			return database.GetUserFromRefreshTokenRow{
				UserID:    userID,
				ExpiresAt: time.Now().UTC().Add(time.Minute),
				RevokedAt: sql.NullTime{Valid: false},
			}, nil
		},
	}
	cfg := &apiConfig{
		tokenSecret: "test-secret",
		tokenStore:  db,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	rec := httptest.NewRecorder()

	cfg.handlerRefresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handlerRefresh() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("handlerRefresh() decode error = %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("handlerRefresh() returned empty token")
	}

	gotID, err := auth.ValidateJWT(payload.Token, cfg.tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if gotID != userID {
		t.Fatalf("ValidateJWT() got %v, want %v", gotID, userID)
	}
}

func TestHandlerRefreshExpired(t *testing.T) {
	db := &stubDB{
		getUserFromRefreshToken: func(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error) {
			return database.GetUserFromRefreshTokenRow{
				UserID:    uuid.New(),
				ExpiresAt: time.Now().UTC().Add(-time.Minute),
				RevokedAt: sql.NullTime{Valid: false},
			}, nil
		},
	}
	cfg := &apiConfig{
		tokenSecret: "test-secret",
		tokenStore:  db,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	rec := httptest.NewRecorder()

	cfg.handlerRefresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("handlerRefresh() status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandlerRefreshRevoked(t *testing.T) {
	db := &stubDB{
		getUserFromRefreshToken: func(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error) {
			return database.GetUserFromRefreshTokenRow{
				UserID:    uuid.New(),
				ExpiresAt: time.Now().UTC().Add(time.Minute),
				RevokedAt: sql.NullTime{Valid: true, Time: time.Now().UTC()},
			}, nil
		},
	}
	cfg := &apiConfig{
		tokenSecret: "test-secret",
		tokenStore:  db,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	rec := httptest.NewRecorder()

	cfg.handlerRefresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("handlerRefresh() status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandlerRevokeSuccess(t *testing.T) {
	called := false
	db := &stubDB{
		getUserFromRefreshToken: func(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error) {
			return database.GetUserFromRefreshTokenRow{
				UserID:    uuid.New(),
				ExpiresAt: time.Now().UTC().Add(time.Minute),
				RevokedAt: sql.NullTime{Valid: false},
			}, nil
		},
		revokeRefreshToken: func(ctx context.Context, token string) error {
			if token != "refresh-token" {
				return errors.New("unexpected token")
			}
			called = true
			return nil
		},
	}
	cfg := &apiConfig{
		tokenSecret: "test-secret",
		tokenStore:  db,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	rec := httptest.NewRecorder()

	cfg.handlerRevoke(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("handlerRevoke() status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatalf("handlerRevoke() did not revoke token")
	}
}

func TestHandlerRevokeMissingToken(t *testing.T) {
	db := &stubDB{
		getUserFromRefreshToken: func(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error) {
			return database.GetUserFromRefreshTokenRow{}, errors.New("not found")
		},
	}
	cfg := &apiConfig{
		tokenSecret: "test-secret",
		tokenStore:  db,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	rec := httptest.NewRecorder()

	cfg.handlerRevoke(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("handlerRevoke() status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
