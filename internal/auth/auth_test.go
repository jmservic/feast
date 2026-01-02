package auth

import (
	"testing"
	//	"github.com/google/uuid"
	//	"time"
	//	"net/http"
)

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
