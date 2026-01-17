package auth

import (
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

func TestFailing(t *testing.T) {
	t.Error("Failing just because lol")
}

func TestHashPasswordReturnsHashString(t *testing.T) {
	password := "password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Error hashing the password: %s", err)
	}
	if hash == "" {
		t.Error("Hash is an empty string...")
	}
	if hash == password {
		t.Error("Hash is equal to the original password.")
	}
}

func TestHashPasswordReturnsUniqueHashes(t *testing.T) {
	password1 := "password"
	password2 := "drowssap"
	hash1, err := HashPassword(password1)
	if err != nil {
		t.Error(err)
	}
	hash2, err := HashPassword(password2)
	if err != nil {
		t.Error(err)
	}
	if hash1 == hash2 {
		t.Error("Hashes are not unique.")
	}
}

func TestCheckPasswordHashReturnsTrueForCorrectPassword(t *testing.T) {
	password := "Passw0rd"
	hash, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Error(err)
	}
	if !match {
		t.Error("password and hash do not match.")
	}
}

func TestCheckPasswordHashReturnsFalseForIncorrectPassword(t *testing.T) {
	password := "Passw0rd"
	hash, err := HashPassword(password)
	if err != nil {
		t.Error(err)
		return
	}
	match, err := CheckPasswordHash("password", hash)
	if err != nil {
		t.Error(err)
		return
	}
	if match {
		t.Error("incorrect password matches hash.")
	}

}

func TestJWTValidation(t *testing.T) {
	tokenSecret := "iovelyxd"
	userID := uuid.New()
	expiresIn := 5 * time.Second
	validToken, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: tokenSecret,
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: tokenSecret,
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, want err %v", err, tt.wantErr)
				return
			}
			if tt.wantUserID != parsedUserID {
				t.Errorf("ValidateJWT() parsedUserID = %v, want %v", parsedUserID, tt.wantUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		header      http.Header
		tokenString string
		wantErr     bool
	}{
		{
			name: "Valid Authorization Header",
			header: http.Header{
				"Authorization": []string{"Bearer abcdefghijklmnopqrstuvwxyz"},
			},
			tokenString: "abcdefghijklmnopqrstuvwxyz",
			wantErr:     false,
		},
		{
			name: "Missing header",
			header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			tokenString: "",
			wantErr:     true,
		},
		{
			name: "Incorrectly formatted header",
			header: http.Header{
				"Authorization": []string{"abcdefghijkmnopqrstuvwxyz!"},
			},
			tokenString: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := GetBearerToken(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, want err %v", err, tt.wantErr)
				return
			}
			if tokenString != tt.tokenString {
				t.Errorf("GetBearerToken tokenString = %v, want %v", tokenString, tt.tokenString)
			}
		})
	}
}
