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

type UserInfo struct {
	name     string
	email    string
	password string
}

// TO-DO: Add failing test cases like a bad name, email, or password. Also different email casing
func TestCreateNewUser(t *testing.T) {
	helpers.LoadDotEnv()
	name := "jonathan"
	email := "Jon@examPle.com"
	password := "very-secret!"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

	res := dto.CreateUser(t, feastUrl, name, email, password)
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
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

	res := dto.CreateUser(t, feastUrl, name, email, password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	sut := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&sut, res.Body, t)

	dto.ValidateUserCreateResponse(t, sut, name, email)

	res.Body.Close()
	res = dto.CreateUser(t, feastUrl, name, email, password)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status bad request, got: %d", res.StatusCode)
	}
}

func TestUserLogin(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

	name := "jonathan"
	email := "jon@example.com"
	password := "very-secret!"

	//Create the user
	res := dto.CreateUser(t, feastUrl, name, email, password)
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
			res := dto.LoginUser(t, feastUrl, testCase.email, testCase.password)
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
	res := dto.CreateUser(t, feastUrl, firstUser.name, firstUser.email, firstUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, firstUser.name, firstUser.email)

	res.Body.Close()

	//second user
	res = dto.CreateUser(t, feastUrl, secondUser.name, secondUser.email, secondUser.password)

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
			res := dto.LoginUser(t, feastUrl, testCase.userInfo.email, testCase.userInfo.password)
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
			res = dto.RefreshUser(t, feastUrl, userLoginResponse.RefreshToken)

			if (res.StatusCode != http.StatusOK && !refreshShouldFail) || (res.StatusCode != http.StatusUnauthorized && refreshShouldFail) {
				t.Fatalf("Refresh should fail = %v, yet received a status code of %d", refreshShouldFail, res.StatusCode)

			}
			res.Body.Close()

			//Test login!
			res = dto.LoginUser(t, feastUrl, testCase.userInfo.email, testCase.userInfo.password)
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Expected an OK response code, received: %d", res.StatusCode)
			}
			res.Body.Close()
		})
	}

}

func TestDeleteUser(t *testing.T) {
	helpers.LoadDotEnv()
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

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
	res := dto.CreateUser(t, feastUrl, firstUser.name, firstUser.email, firstUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, firstUser.name, firstUser.email)

	res.Body.Close()

	//second user
	res = dto.CreateUser(t, feastUrl, secondUser.name, secondUser.email, secondUser.password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse = dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)

	dto.ValidateUserCreateResponse(t, userCreateResponse, secondUser.name, secondUser.email)
	res.Body.Close()

	//Login for first user
	res = dto.LoginUser(t, feastUrl, firstUser.email, firstUser.password)
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
			res = dto.LoginUser(t, feastUrl, testCase.userInfo.email, testCase.userInfo.password)
			res.Body.Close()
			if (testCase.responseCode == http.StatusOK && res.StatusCode != http.StatusUnauthorized) ||
				(testCase.responseCode == http.StatusBadRequest && res.StatusCode != http.StatusOK) {
				t.Fatalf("Delete expected response code = %d, but received a login response code of %d", testCase.responseCode, res.StatusCode)
			}

			// Test Refresh if we have a token
			if testCase.refreshToken != "" {
				res = dto.RefreshUser(t, feastUrl, testCase.refreshToken)
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
