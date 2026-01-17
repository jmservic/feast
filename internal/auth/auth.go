package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "feast-access"
)

// TO-DO: Review hashing method for the server
var argon2idParams = argon2id.DefaultParams

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2idParams)
	if err != nil {
		return "", fmt.Errorf("error creating hash from %s: %w", password, err)
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("error comparing %s and %s: %w", password, hash, err)
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(*jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		})
	if err != nil {
		fmt.Println(tokenString)
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}
	if err != nil {
		return uuid.Nil, errors.New("error getting jwt issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	authFields := strings.Fields(authHeader)
	if len(authFields) != 2 {
		return "", errors.New("authorization header is in the wrong format")
	}
	return authFields[1], nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	authFields := strings.Fields(authHeader)
	if len(authFields) != 2 {
		return "", errors.New("authorization header is in the wrong format")
	}
	return authFields[1], nil
}

func MakeRefreshToken() (string, error) {
	refreshBytes := make([]byte, 32)
	rand.Read(refreshBytes)
	return hex.EncodeToString(refreshBytes), nil
}
