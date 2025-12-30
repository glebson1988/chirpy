package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatalf("HashPassword() returned empty hash")
	}

	ok, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash() error = %v", err)
	}
	if !ok {
		t.Fatalf("CheckPasswordHash() expected true for matching password")
	}

	ok, err = CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash() error = %v", err)
	}
	if ok {
		t.Fatalf("CheckPasswordHash() expected false for non-matching password")
	}
}

func TestJWTCreateAndValidate(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if gotID != userID {
		t.Fatalf("ValidateJWT() got %v, want %v", gotID, userID)
	}
}

func TestJWTExpiredTokenRejected(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, -time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	if _, err := ValidateJWT(token, secret); err == nil {
		t.Fatalf("ValidateJWT() expected error for expired token")
	}
}

func TestJWTWrongSecretRejected(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	if _, err := ValidateJWT(token, "other-secret"); err == nil {
		t.Fatalf("ValidateJWT() expected error for wrong secret")
	}
}
