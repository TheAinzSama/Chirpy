package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWTAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "supersecretkey"
	expiresIn := time.Minute

	// Test creating a valid token
	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	// Test validating the token
	validatedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if validatedID != userID {
		t.Errorf("Expected userID %s, got %s", userID, validatedID)
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer my_access_token_123")
	token, err := GetBearerToken(headers)
	println(token)
	if err != nil {
		t.Errorf("GetBearerToken returned an error: %v", err)
	}

	if token != "my_access_token_123" {
		t.Errorf("GetBearerToken returned incorrect token: got %s, want %s", token, "my_access_token_123")
	}

}
