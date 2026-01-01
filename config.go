package main

import (
	"context"
	"sync/atomic"

	"github.com/glebson1988/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	tokenSecret    string
	tokenStore     tokenStore
	userStore      userStore
}

type tokenStore interface {
	GetUserFromRefreshToken(ctx context.Context, token string) (database.GetUserFromRefreshTokenRow, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type userStore interface {
	SetChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error)
}
