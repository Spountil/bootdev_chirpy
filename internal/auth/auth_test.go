package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHash(t *testing.T) {
	tests := []struct {
		password string
	}{
		{
			password: "test123",
		},
		{
			password: "TESSSSST7896",
		},
		{
			password: "HelloWorld",
		},
		{
			password: "INeedItToWork",
		},
		{
			password: "HelpMe987..=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("got error %v when hashing on password %s", err, tt.password)
			}

			isSame, err := CheckPasswordHash(tt.password, hashedPassword)
			if err != nil {
				t.Fatalf("got error %v on comparing password %s", err, tt.password)
			}
			if !isSame {
				return
			}
		})
	}
}

func TestJWT(t *testing.T) {
	secret := "my-super-secret-key"
	userID := uuid.New()

	t.Run("Valid Token", func(t *testing.T) {
		token, err := MakeJWT(userID, secret, time.Hour)
		if err != nil {
			t.Fatalf("Expected no error creating JWT, got %v", err)
		}

		parsedID, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("Expected no error validating JWT, got %v", err)
		}

		if parsedID != userID {
			t.Errorf("Expected parsed user ID to be %v, got %v", userID, parsedID)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		token, err := MakeJWT(userID, secret, -time.Hour)
		if err != nil {
			t.Fatalf("Expected no error creating JWT, got %v", err)
		}

		_, err = ValidateJWT(token, secret)
		if err == nil {
			t.Error("Expected an error validating an expired token, got nil")
		}
	})

	t.Run("Wrong Secret", func(t *testing.T) {
		wrongSecret := "definitely-not-the-right-secret"

		token, err := MakeJWT(userID, secret, time.Hour)
		if err != nil {
			t.Fatalf("Expected no error creating JWT, got %v", err)
		}

		_, err = ValidateJWT(token, wrongSecret)
		if err == nil {
			t.Error("Expected an error validating token with wrong secret, got nil")
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectErr     bool
	}{
		{
			name: "Valid standard Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer my-secret-token-123"},
			},
			expectedToken: "my-secret-token-123",
			expectErr:     false,
		},
		{
			name: "Valid token with excessive spacing",
			headers: http.Header{
				"Authorization": []string{"  Bearer    token-with-spaces  "},
			},
			expectedToken: "token-with-spaces",
			expectErr:     false,
		},
		{
			name:          "Missing Authorization header entirely",
			headers:       http.Header{},
			expectedToken: "",
			expectErr:     true,
		},
		{
			name: "Empty Authorization header value",
			headers: http.Header{
				"Authorization": []string{""},
			},
			expectedToken: "",
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.expectErr {
				t.Fatalf("GetBearerToken() error = %v, expectErr %v", err, tt.expectErr)
			}

			if token != tt.expectedToken {
				t.Errorf("GetBearerToken() got Token = %q, want %q", token, tt.expectedToken)
			}
		})
	}
}
