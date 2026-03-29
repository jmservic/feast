package constants

import (
	"time"
)

const (
	EmptyPasswordErrStr            string = "password field cannot be empty"
	EmptyParameterErrStr           string = "a required parameter is empty"
	HashCheckErrStr                string = "error comparing hashes"
	HouseholdCreationErrStr        string = "error creating a new household"
	HouseholdDeleteErrStr          string = "error deleting the household"
	HouseholdRetrievalByUserErrStr string = "error retrieving household by user id"
	HouseholdRetrievalByIdErrStr   string = "error retrieving household by id"
	HouseholdUpdateErrStr          string = "error updating a household"
	InvalidCredentialsErrStr       string = "invalid credentials" // #nosec G101
	JsonDecodeErrStr               string = "error occurred when decoding the json string"
	JwtCreationErrStr              string = "error creating a JWT token"
	JwtRetrievalErrStr             string = "error getting JWT"
	JwtValidationErrStr            string = "error validating JWT"
	PasswordHashErrStr             string = "error hashing the password" // #nosec G101
	RefreshTokenCheckingErrStr     string = "error checking refresh token"
	RefreshTokenCreationErrStr     string = "error creating a refresh token"
	RefreshTokenInvalidErrStr      string = "invalid refresh token"
	RefreshTokenRetrievalErrStr    string = "error getting refresh token"
	RefreshTokenRevokeErrStr       string = "error revoking refresh token"
	RefreshTokenStorageErrStr      string = "error storing a refresh token"
	UserCreationErrStr             string = "error creating a new user"
	UserIdRetrievalErrStr          string = "error getting user id"
	UserUpdateErrStr               string = "error updating user information"
	UserDeleteErrStr               string = "error deleting the user"
)

const (
	AccessTokenLength  time.Duration = time.Minute * 15
	RefreshTokenLength time.Duration = time.Hour * 24 * 60
)
