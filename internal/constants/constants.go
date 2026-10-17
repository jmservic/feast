package constants

import (
	"time"
)

const (
	BeginTransactionErrStr                       string = "error creating a database transaction"
	CommitTransactionErrStr                      string = "error commiting the database transaction"
	EmptyPasswordErrStr                          string = "password field cannot be empty"
	EmptyParameterErrStr                         string = "a required parameter is empty"
	GetIngredientErrStr                          string = "error getting ingredient by id"
	GetIngredientsErrStr                         string = "error getting ingredients by query"
	GetIngredientServingsErrStr                  string = "error getting ingredient servings"
	GetIngredientsServingsErrStr                 string = "error getting ingredient servings by ids"
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
	IngredientCreationErrStr                     string = "error creating a new ingredient"
	InvalidCalorieErrStr                         string = "calorie value cannot be negative"
	InvalidCarbohydratesErrStr                   string = "carbohydrates  value cannot be negative"
	InvalidFatsErrStr                            string = "fats value cannot be negative"
	InvalidIntStringErrStr                       string = "error converting string to intger"
	InvalidProteinErrStr                         string = "protein value cannot be negative"
	InvalidAmountErrStr                          string = "amount value must be greater than zero"
	InvalidCredentialsErrStr                     string = "invalid credentials" // #nosec G101
	InvalidMultipleServingsErrStr                string = "serving quantity must be greater than zero for multiple servings"
	InvalidServingsErrStr                        string = "serving quantity cannot be negative"
	InvalidUUIDErrStr                            string = "error parsing uuid string"
	InviteAcceptErrStr                           string = "error accepting household invitation"
	InviteDeclineErrStr                          string = "error declining household invitation"
	JsonDecodeErrStr                             string = "error occurred when decoding the json string"
	JwtCreationErrStr                            string = "error creating a JWT token"
	JwtRetrievalErrStr                           string = "error getting JWT"
	JwtValidationErrStr                          string = "error validating JWT"
	LeaveHouseholdErrStr                         string = "error leaving your household"
	NotAuthorizedErrStr                          string = "not authorized"
	PasswordHashErrStr                           string = "error hashing the password" // #nosec G101
	PromoteHouseholdMemberToHeadErrStr           string = "error promoting the household member to the household head"
	RefreshTokenCheckingErrStr                   string = "error checking refresh token"
	RefreshTokenCreationErrStr                   string = "error creating a refresh token"
	RefreshTokenInvalidErrStr                    string = "invalid refresh token"
	RefreshTokenRetrievalErrStr                  string = "error getting refresh token"
	RefreshTokenRevokeErrStr                     string = "error revoking refresh token"
	RefreshTokenStorageErrStr                    string = "error storing a refresh token"
	ServingsCreationErrStr                       string = "error creating list of servings for an ingredient"
	UserCreationErrStr                           string = "error creating a new user"
	UserIdRetrievalErrStr                        string = "error getting user id"
	UserUpdateErrStr                             string = "error updating user information"
	UserDeleteErrStr                             string = "error deleting the user"
)

const (
	AccessTokenLength  time.Duration = time.Minute * 15
	RefreshTokenLength time.Duration = time.Hour * 24 * 60
)
