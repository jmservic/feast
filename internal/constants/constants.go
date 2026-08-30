package constants

import (
	"time"
)

const (
	BeginTransactionErrStr                       string = "error creating a database transaction"
	EmptyPasswordErrStr                          string = "password field cannot be empty"
	EmptyParameterErrStr                         string = "a required parameter is empty"
	GetNewInstanceErrStr                         string = "error getting the newly created instance"
	HashCheckErrStr                              string = "error comparing hashes"
	HandleMemberInviteActionErrStr               string = "Invalid string representing invite action, must be either accept or decline"
	HouseholdCreationErrStr                      string = "error creating a new household"
	HouseholdDeleteErrStr                        string = "error deleting the household"
	HouseholdMemberCreateErrStr                  string = "error creating a new household member"
	HouseholdMemberDeleteErrStr                  string = "error deleting the household member"
	HouseholdMemberInviteErrStr                  string = "error inviting user to the household"
	HouseholdMemberRetrievalByIdErrStr           string = "error retrieving the household member by id"
	HouseholdMembersRetrievalByHouseholdIdErrStr string = "error retrieving the household members"
	HouseholdMemberUpdateErrStr                  string = "error updating the household member"
	HouseholdRetrievalByUserErrStr               string = "error retrieving the household by user id"
	HouseholdRetrievalByIdErrStr                 string = "error retrieving the household by id"
	HouseholdUpdateErrStr                        string = "error updating a household"
	MemberInviteRetrievalByUserErrStr            string = "error getting member invites by user ID"
	InvalidCredentialsErrStr                     string = "invalid credentials" // #nosec G101
	InvalidUUIDErrStr                            string = "error parsing uuid string"
	InviteAcceptErrStr                           string = "error accepting household invitation"
	InviteDeclineErrStr                          string = "error declining household invitation"
	JsonDecodeErrStr                             string = "error occurred when decoding the json string"
	JwtCreationErrStr                            string = "error creating a JWT token"
	JwtRetrievalErrStr                           string = "error getting JWT"
	JwtValidationErrStr                          string = "error validating JWT"
	PasswordHashErrStr                           string = "error hashing the password" // #nosec G101
	RefreshTokenCheckingErrStr                   string = "error checking refresh token"
	RefreshTokenCreationErrStr                   string = "error creating a refresh token"
	RefreshTokenInvalidErrStr                    string = "invalid refresh token"
	RefreshTokenRetrievalErrStr                  string = "error getting refresh token"
	RefreshTokenRevokeErrStr                     string = "error revoking refresh token"
	RefreshTokenStorageErrStr                    string = "error storing a refresh token"
	UserCreationErrStr                           string = "error creating a new user"
	UserIdRetrievalErrStr                        string = "error getting user id"
	UserUpdateErrStr                             string = "error updating user information"
	UserDeleteErrStr                             string = "error deleting the user"
)

const (
	AccessTokenLength  time.Duration = time.Minute * 15
	RefreshTokenLength time.Duration = time.Hour * 24 * 60
)
