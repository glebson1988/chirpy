package main

import (
	"sync/atomic"

	"github.com/glebson1988/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	tokenSecret    string
}
