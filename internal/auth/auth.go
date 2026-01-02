package auth

import (
	"fmt"
	"github.com/alexedwards/argon2id"
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
