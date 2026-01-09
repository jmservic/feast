package constants

import (
	"time"
)

const (
	EmptyPasswordErrStr         string = "password field cannot be empty"
	HashCheckErrStr             string = "error comparing hashes"
	InvalidCredentialsErrStr    string = "invalid credentials"
	JsonDecodeErrStr            string = "error occurred when decoding the json string"
	JwtCreationErrStr           string = "error creating a JWT token"
	JwtRetrievalErrStr          string = "error getting JWT"
	JwtValidationErrStr         string = "error validating JWT"
	PasswordHashErrStr          string = "error hashing the password"
	RefreshTokenCheckingErrStr  string = "error checking refresh token"
	RefreshTokenCreationErrStr  string = "error creating a refresh token"
	RefreshTokenInvalidErrStr   string = "invalid refresh token"
	RefreshTokenRetrievalErrStr string = "error getting refresh token"
	RefreshTokenRevokeErrStr    string = "error revoking refresh token"
	RefreshTokenStorageErrStr   string = "error storing a refresh token"
	UserCreationErrStr          string = "error creating a new user"
	UserIdRetrievalErrStr       string = "error getting user id"
)

const (
	AccessTokenLength  time.Duration = time.Minute * 15
	RefreshTokenLength time.Duration = time.Hour * 24 * 60
)
