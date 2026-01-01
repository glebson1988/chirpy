package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebson1988/chirpy/internal/database"
	"github.com/google/uuid"
)

type stubUserStore struct {
	setChirpyRed func(ctx context.Context, id uuid.UUID) (database.User, error)
}

func (s *stubUserStore) SetChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error) {
	if s.setChirpyRed == nil {
		return database.User{}, errors.New("not implemented")
	}
	return s.setChirpyRed(ctx, id)
}

func TestHandlerPolkaWebhooksIgnoresOtherEvents(t *testing.T) {
	cfg := &apiConfig{
		userStore: &stubUserStore{},
	}

	payload := polkaWebhookRequest{
		Event: "user.created",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	cfg.handlerPolkaWebhooks(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("handlerPolkaWebhooks() status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestHandlerPolkaWebhooksUpgradedSuccess(t *testing.T) {
	userID := uuid.New()
	cfg := &apiConfig{
		userStore: &stubUserStore{
			setChirpyRed: func(ctx context.Context, id uuid.UUID) (database.User, error) {
				if id != userID {
					return database.User{}, errors.New("unexpected id")
				}
				return database.User{ID: id}, nil
			},
		},
	}

	payload := polkaWebhookRequest{
		Event: "user.upgraded",
		Data: struct {
			UserID string `json:"user_id"`
		}{UserID: userID.String()},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	cfg.handlerPolkaWebhooks(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("handlerPolkaWebhooks() status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestHandlerPolkaWebhooksUserNotFound(t *testing.T) {
	userID := uuid.New()
	cfg := &apiConfig{
		userStore: &stubUserStore{
			setChirpyRed: func(ctx context.Context, id uuid.UUID) (database.User, error) {
				return database.User{}, sql.ErrNoRows
			},
		},
	}

	payload := polkaWebhookRequest{
		Event: "user.upgraded",
		Data: struct {
			UserID string `json:"user_id"`
		}{UserID: userID.String()},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	cfg.handlerPolkaWebhooks(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("handlerPolkaWebhooks() status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerPolkaWebhooksInvalidUserID(t *testing.T) {
	cfg := &apiConfig{
		userStore: &stubUserStore{},
	}

	payload := polkaWebhookRequest{
		Event: "user.upgraded",
		Data: struct {
			UserID string `json:"user_id"`
		}{UserID: "not-a-uuid"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	cfg.handlerPolkaWebhooks(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("handlerPolkaWebhooks() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
