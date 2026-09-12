package integration

import (
	//"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	//"github.com/jmservic/feast/integration_tests/constants"
	"io"
	"net/http"
	//"os"
	"strings"
	"testing"
)

// TO-DO: Add failing test cases like a bad name, email, or password. Also different email casing
func TestCreateNewUser(t *testing.T) {
	helpers.LoadDotEnv()
	name := "jonathan"
	email := "Jon@examPle.com"
	password := "very-secret!"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	res := client.CreateUser(name, email, password)
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}

	sut := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&sut, res.Body, t)

	dto.ValidateUserCreateResponse(t, sut, name, email)
}

func TestCreateDuplicateUserFails(t *testing.T) {
	helpers.LoadDotEnv()
	name := "jonathan"
	email := "jon@example.com"
	password := "very-secret!"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	res := client.CreateUser(name, email, password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	sut := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&sut, res.Body, t)

	dto.ValidateUserCreateResponse(t, sut, name, email)

	res.Body.Close()
	res = client.CreateUser(name, email, password)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status bad request, got: %d", res.StatusCode)
	}
}

func TestUserLogin(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	name := "jonathan"
	email := "jon@example.com"
	password := "very-secret!"

	//Create the user
	res := client.CreateUser(name, email, password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, name, email)

	res.Body.Close()

	//Authenticate the user
	testCases := []struct {
		email        string
		password     string
		responseCode int
		testName     string
	}{
		{
			email:        email,
			password:     password,
			responseCode: http.StatusOK,
			testName:     "Correct Credentials",
		},
		{
			email:        email,
			password:     "wrong-passw0rd",
			responseCode: http.StatusUnauthorized,
			testName:     "Incorrect Password",
		},
		{
			email:        "user@example.com",
			password:     password,
			responseCode: http.StatusUnauthorized,
			testName:     "Incorrect Email",
		},
		{
			email:        strings.ToUpper(email),
			password:     password,
			responseCode: http.StatusOK,
			testName:     "Correct Credentials with different email casing.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.LoginUser(testCase.email, testCase.password)
			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected status code %d, got: %d", testCase.responseCode, res.StatusCode)
			}

			switch res.StatusCode {
			case http.StatusOK:
				sut := dto.UserLoginResponse{}
				helpers.DecodeJSONResponse(&sut, res.Body, t)

				if sut.Email != strings.ToLower(testCase.email) {
					t.Fatal("Payload and response emails do not match")
				}
				if len(sut.Token) == 0 {
					t.Fatal("Received an empty access token")
				}
				if len(sut.RefreshToken) == 0 {
					t.Fatal("Received an empty refresh token")
				}
			default:
				return
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	firstUser := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	secondUser := UserInfo{
		name:     "cassidy",
		email:    "cassidy@example.com",
		password: "kalina",
	}

	// Create the two users
	//first user
	res := client.CreateUser(firstUser.name, firstUser.email, firstUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, firstUser.name, firstUser.email)

	res.Body.Close()

	//second user
	res = client.CreateUser(secondUser.name, secondUser.email, secondUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse = dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, secondUser.name, secondUser.email)
	res.Body.Close()

	testCases := []struct {
		userInfo     *UserInfo
		newName      string
		newEmail     string
		newPassword  string
		responseCode int
		testName     string
	}{
		{
			userInfo:     &firstUser,
			newName:      "Jonathan Service",
			newEmail:     firstUser.email,
			newPassword:  firstUser.password,
			responseCode: http.StatusOK,
			testName:     "New User Name",
		},
		{
			userInfo:     &firstUser,
			newName:      "Jonathan Service",
			newEmail:     secondUser.email,
			newPassword:  "different-password",
			responseCode: http.StatusBadRequest,
			testName:     "Updating to already in use email",
		},
		{
			userInfo:     &secondUser,
			newName:      secondUser.name,
			newEmail:     secondUser.email,
			newPassword:  "bobina",
			responseCode: http.StatusOK,
			testName:     "New Password",
		},
		{
			userInfo:     &secondUser,
			newName:      secondUser.name,
			newEmail:     "castadon@example.com",
			newPassword:  "bobina",
			responseCode: http.StatusOK,
			testName:     "New Email",
		},
		{
			userInfo:     &firstUser,
			newName:      "Jonathan Service",
			newEmail:     "Inqindi@example.com",
			newPassword:  "axel&brie&cindy&kalina",
			responseCode: http.StatusOK,
			testName:     "New Email and Password",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			refreshShouldFail := false
			//Login
			res := client.LoginUser(testCase.userInfo.email, testCase.userInfo.password)
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Expected an OK response code, received: %d", res.StatusCode)
			}

			userLoginResponse := dto.UserLoginResponse{}
			helpers.DecodeJSONResponse(&userLoginResponse, res.Body, t)
			res.Body.Close()

			// Update User
			res = dto.UpdateUser(t, feastUrl, userLoginResponse.Token, testCase.newName, testCase.newEmail, testCase.newPassword)
			switch res.StatusCode {
			case http.StatusOK:
				//If the Email or Password has changed, the refresh should fail.
				//Update the user
				if testCase.userInfo.name != testCase.newName {
					testCase.userInfo.name = testCase.newName
				}

				if testCase.userInfo.email != testCase.newEmail {
					testCase.userInfo.email = testCase.newEmail
					refreshShouldFail = true
				}

				if testCase.userInfo.password != testCase.newPassword {
					testCase.userInfo.password = testCase.newPassword
					refreshShouldFail = true
				}
			default:
				//No-op
			}

			if res.StatusCode != testCase.responseCode {
				buffer, _ := io.ReadAll(res.Body)
				t.Fatalf("Expected an %d response code, received: %d: %v", testCase.responseCode, res.StatusCode, string(buffer))
			}
			res.Body.Close()

			//Test refresh
			res = client.RefreshUser(userLoginResponse.RefreshToken)

			if (res.StatusCode != http.StatusOK && !refreshShouldFail) || (res.StatusCode != http.StatusUnauthorized && refreshShouldFail) {
				t.Fatalf("Refresh should fail = %v, yet received a status code of %d", refreshShouldFail, res.StatusCode)

			}
			res.Body.Close()

			//Test login!
			res = client.LoginUser(testCase.userInfo.email, testCase.userInfo.password)
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Expected an OK response code, received: %d", res.StatusCode)
			}
			res.Body.Close()
		})
	}

}

// Test for when we delete the user.... who is the owner of a household, and also a member of a household. The member id should be cleared
func TestDeleteUser(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	firstUser := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	secondUser := UserInfo{
		name:     "cassidy",
		email:    "cassidy@example.com",
		password: "kalina",
	}

	// Create the two users
	//first user
	res := client.CreateUser(firstUser.name, firstUser.email, firstUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, firstUser.name, firstUser.email)

	res.Body.Close()

	//second user
	res = client.CreateUser(secondUser.name, secondUser.email, secondUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse = dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, secondUser.name, secondUser.email)
	res.Body.Close()

	//Login for first user
	res = client.LoginUser(firstUser.email, firstUser.password)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status ok for user login, got : %d", res.StatusCode)
	}
	userLoginResponse := dto.UserLoginResponse{}
	helpers.DecodeJSONResponse(&userLoginResponse, res.Body, t)
	res.Body.Close()

	testCases := []struct {
		userInfo                      *UserInfo
		token, refreshToken, testName string
		responseCode                  int
	}{
		{
			userInfo:     &firstUser,
			token:        userLoginResponse.Token,
			refreshToken: userLoginResponse.RefreshToken,
			testName:     "Successful",
			responseCode: http.StatusOK,
		},
		{
			userInfo:     &secondUser,
			token:        "bad-token",
			refreshToken: "",
			testName:     "Failure",
			responseCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			// Test Delete
			res := dto.DeleteUser(t, feastUrl, testCase.token)
			if res.StatusCode != testCase.responseCode {
				t.Errorf("Expected status code of %d for the delete request, got : %d", testCase.responseCode, res.StatusCode)
			}
			res.Body.Close()

			// Test Login
			res = client.LoginUser(testCase.userInfo.email, testCase.userInfo.password)
			res.Body.Close()
			if (testCase.responseCode == http.StatusOK && res.StatusCode != http.StatusUnauthorized) ||
				(testCase.responseCode == http.StatusBadRequest && res.StatusCode != http.StatusOK) {
				t.Fatalf("Delete expected response code = %d, but received a login response code of %d", testCase.responseCode, res.StatusCode)
			}

			// Test Refresh if we have a token
			if testCase.refreshToken != "" {
				res = client.RefreshUser(testCase.refreshToken)
				//If the delete was successful this should fail, else it should succeed.
				if (testCase.responseCode == http.StatusOK && res.StatusCode != http.StatusUnauthorized) ||
					(testCase.responseCode == http.StatusBadRequest && res.StatusCode != http.StatusOK) {
					t.Fatalf("Delet expected response code = %d, but received a refresh token response code of %d", testCase.responseCode, res.StatusCode)
				}
				res.Body.Close()
			}
		})
	}
}

func TestDeleteUsersHouseholdOwnerWithNoMembers(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	householdName := "Service Family"
	owner := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	// create the owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	//login Owner
	res = client.LoginUser(owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = dto.DeleteUser(t, feastUrl, ownerTokenRes.Token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected delete user to return status OK, but received: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(owner.email, owner.password)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected the login user to return status unauthorized, but received: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.GetHousehold(householdInfo.Id.String(), ownerTokenRes.Token)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected get household to return status not found, received: %d", res.StatusCode)
	}
	res.Body.Close()
}

func TestDeleteUsersHouseholdOwnerWithMembers(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	householdName := "Service Family"
	owner := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	secondMember := UserInfo{
		name:     "joey",
		email:    "secondLyphe@example.com",
		password: "face-pain",
	}

	// create the owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}

	res = client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(secondMember.name, secondMember.email, secondMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// login as users
	//Owner
	res = client.LoginUser(owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	res = client.LoginUser(member.email, member.password)
	memberLoginRes := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(secondMember.email, secondMember.password)
	secondMemberLoginRes := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//Invite, accept and promote members
	res = client.CreateHouseholdMember(ownerTokenRes.Token, member.name, householdInfo.Id, &memberLoginRes.Id)
	memberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.CreateHouseholdMember(ownerTokenRes.Token, secondMember.name, householdInfo.Id, &secondMemberLoginRes.Id)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected the create household member to return statue created, received %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(memberLoginRes.Token, householdInfo.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(secondMemberLoginRes.Token, householdInfo.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(ownerTokenRes.Token, memberInfo.Id, memberLoginRes.Id,
		memberInfo.Name, 2)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the update household member to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	res = dto.DeleteUser(t, feastUrl, ownerTokenRes.Token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected delete user to return status OK, but received: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(owner.email, owner.password)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected the login user to return status unauthorized, but received: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.GetHouseholdMember(memberLoginRes.Token, memberInfo.Id)
	updatedMemberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)
	if updatedMemberInfo.Role != 1 {
		t.Fatalf("Expected the administrator user to now be the household head, but role is %d", updatedMemberInfo.Role)
	}
}

func TestDeleteUsersHouseholdMember(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	householdName := "Service Family"
	owner := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	// create the owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}

	res = client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// login as users
	//Owner
	res = client.LoginUser(owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	res = client.LoginUser(member.email, member.password)
	memberLoginRes := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//Invite and accept member
	res = client.CreateHouseholdMember(ownerTokenRes.Token, member.name, householdInfo.Id, &memberLoginRes.Id)
	memberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.HandleInvite(memberLoginRes.Token, householdInfo.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	res = dto.DeleteUser(t, feastUrl, memberLoginRes.Token)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected delete user to return status OK, but received: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(member.email, member.password)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected the login user to return status unauthorized, but received: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.GetHouseholdMember(memberLoginRes.Token, memberInfo.Id)
	updatedMemberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)
	if updatedMemberInfo.UserId != nil {
		t.Fatal("Expected the household member to have a nil user id")
	}
}
