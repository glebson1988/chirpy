package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer test-token")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken() error = %v", err)
	}
	if token != "test-token" {
		t.Fatalf("GetBearerToken() got %q, want %q", token, "test-token")
	}
}

func TestGetBearerTokenMissingHeader(t *testing.T) {
	headers := http.Header{}

	if _, err := GetBearerToken(headers); err == nil {
		t.Fatalf("GetBearerToken() expected error for missing header")
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("MakeRefreshToken() error = %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("MakeRefreshToken() got length %d, want 64", len(token))
	}
	for _, r := range token {
		isDigit := r >= '0' && r <= '9'
		isLowerHex := r >= 'a' && r <= 'f'
		if !isDigit && !isLowerHex {
			t.Fatalf("MakeRefreshToken() got non-hex character %q", r)
		}
	}
}
