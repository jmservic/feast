package integration

import (
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// TO-DO: Add failing test cases like a bad name, email, or password. Also different email casing
func TestCreateNewUser(t *testing.T) {
	loadDotEnv()
	name := "jonathan"
	email := "Jon@examPle.com"
	password := "very-secret!"

	feast_url := "http://localhost:" + os.Getenv("PORT")
	t.Cleanup(func() { resetDatabase(feast_url) })

	payload := UserCreatePayload{
		Name:     name,
		Email:    email,
		Password: password,
	}

	body := CreateJSONReader(payload, t)

	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	sut := UserCreateResponse{}
	DecodeJSONResponse(&sut, res.Body, t)

	if sut.Name != name {
		t.Fatalf("Expected %s, but got %s for the name", name, sut.Name)
	}
	if sut.Email != strings.ToLower(email) {
		t.Fatalf("Expected %s, but got %s for the email", email, sut.Email)
	}
	if sut.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the user id")
	}

}

func TestCreateDuplicateUserFails(t *testing.T) {
	loadDotEnv()
	name := "jonathan"
	email := "jon@example.com"
	password := "very-secret!"

	feast_url := getFeastURL()
	t.Cleanup(func() { resetDatabase(feast_url) })

	payload := UserCreatePayload{
		Name:     name,
		Email:    email,
		Password: password,
	}

	body := CreateJSONReader(payload, t)

	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	sut := UserCreateResponse{}
	DecodeJSONResponse(&sut, res.Body, t)

	if sut.Name != name {
		t.Fatalf("Expected %s, but got %s for the name", name, sut.Name)
	}
	if sut.Email != email {
		t.Fatalf("Expected %s, but got %s for the email", email, sut.Email)
	}
	if sut.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the user id")
	}

	res.Body.Close()
	body.Seek(0, io.SeekStart)
	res, err = http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status bad request, got: %d", res.StatusCode)
	}
}

func TestUserLogin(t *testing.T) {
	loadDotEnv()
	feast_url := getFeastURL()
	t.Cleanup(func() { resetDatabase(feast_url) })

	name := "jonathan"
	email := "jon@example.com"
	password := "very-secret!"

	//Create the user
	payload := UserCreatePayload{
		Name:     name,
		Email:    email,
		Password: password,
	}

	body := CreateJSONReader(payload, t)
	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := UserCreateResponse{}
	DecodeJSONResponse(&userCreateResponse, res.Body, t)

	if userCreateResponse.Name != name {
		t.Fatalf("Expected %s, but got %s for the name", name, userCreateResponse.Name)
	}
	if userCreateResponse.Email != email {
		t.Fatalf("Expected %s, but got %s for the email", email, userCreateResponse.Email)
	}
	if userCreateResponse.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the user id")
	}

	res.Body.Close()

	//Authenticate the user
	testCases := []struct {
		payload      UserLoginPayload
		responseCode int
		testName     string
	}{
		{
			payload: UserLoginPayload{
				Email:    email,
				Password: password,
			},
			responseCode: http.StatusOK,
			testName:     "Correct Credentials",
		},
		{
			payload: UserLoginPayload{
				Email:    email,
				Password: "wrong-passw0rd",
			},
			responseCode: http.StatusUnauthorized,
			testName:     "Incorrect Password",
		},
		{
			payload: UserLoginPayload{
				Email:    "user@example.com",
				Password: password,
			},
			responseCode: http.StatusUnauthorized,
			testName:     "Incorrect Email",
		},
		{
			payload: UserLoginPayload{
				Email:    strings.ToUpper(email),
				Password: password,
			},
			responseCode: http.StatusOK,
			testName:     "Correct Credentials with different email casing.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			body := CreateJSONReader(testCase.payload, t)
			res, err := http.Post(feast_url+"/api/login", "application/json", body)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected status code %d, got: %d", testCase.responseCode, res.StatusCode)
			}

			switch res.StatusCode {
			case http.StatusOK:
				sut := UserLoginResponse{}
				DecodeJSONResponse(&sut, res.Body, t)

				if sut.Email != strings.ToLower(testCase.payload.Email) {
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
	loadDotEnv()
	feast_url := getFeastURL()
	t.Cleanup(func() { resetDatabase(feast_url) })

	firstUserName := "jonathan"
	firstUserEmail := "jon@example.com"
	firstUserPassword := "very-secret!"

	secondUserName := "cassidy"
	secondUserEmail := "cassidy@example.com"
	secondUserPassword := "kalina"

	// Create the two users
	//first user
	payload := UserCreatePayload{
		Name:     firstUserName,
		Email:    firstUserEmail,
		Password: firstUserPassword,
	}

	body := CreateJSONReader(payload, t)
	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse := UserCreateResponse{}
	DecodeJSONResponse(&userCreateResponse, res.Body, t)

	if userCreateResponse.Name != firstUserName {
		t.Fatalf("Expected %s, but got %s for the name", firstUserName, userCreateResponse.Name)
	}
	if userCreateResponse.Email != firstUserEmail {
		t.Fatalf("Expected %s, but got %s for the email", firstUserEmail, userCreateResponse.Email)
	}
	if userCreateResponse.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the user id")
	}

	res.Body.Close()

	//second user
	payload = UserCreatePayload{
		Name:     secondUserName,
		Email:    secondUserEmail,
		Password: secondUserPassword,
	}

	body = CreateJSONReader(payload, t)
	res, err = http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	userCreateResponse = UserCreateResponse{}
	DecodeJSONResponse(&userCreateResponse, res.Body, t)

	if userCreateResponse.Name != secondUserName {
		t.Fatalf("Expected %s, but got %s for the name", secondUserName, userCreateResponse.Name)
	}
	if userCreateResponse.Email != secondUserEmail {
		t.Fatalf("Expected %s, but got %s for the email", secondUserEmail, userCreateResponse.Email)
	}
	if userCreateResponse.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the user id")
	}

	res.Body.Close()

	testCases := []struct {
		loginPayload UserLoginPayload
		payload      UserUpdatePayload
		responseCode int
		testLogin    bool
		testRefresh  bool
		testName     string
	}{
		{

			loginPayload: UserLoginPayload{
				Email:    firstUserEmail,
				Password: firstUserPassword,
			},
			payload: UserUpdatePayload{
				UserCreatePayload{
					Name:     "Jonathan Service",
					Email:    firstUserEmail,
					Password: firstUserPassword,
				},
			},
			responseCode: http.StatusOK,
			testLogin:    true,
			testRefresh:  true,
			testName:     "New User Name",
		},
		{
			loginPayload: UserLoginPayload{
				Email:    firstUserEmail,
				Password: firstUserPassword,
			},
			payload: UserUpdatePayload{
				UserCreatePayload{
					Name:     "Jonathan Service",
					Email:    secondUserEmail,
					Password: "different-password",
				},
			},
			responseCode: http.StatusBadRequest,
			testLogin:    true,
			testRefresh:  true,
			testName:     "Updating to already in use email",
		},
		{
			loginPayload: UserLoginPayload{
				Email:    secondUserEmail,
				Password: secondUserPassword,
			},
			payload: UserUpdatePayload{
				UserCreatePayload{
					Name:     secondUserName,
					Email:    secondUserEmail,
					Password: "bobina",
				},
			},
			responseCode: http.StatusOK,
			testLogin:    true,
			testRefresh:  true,
			testName:     "New Password",
		},
		{
			loginPayload: UserLoginPayload{
				Email:    secondUserEmail,
				Password: "bobina",
			},
			payload: UserUpdatePayload{
				UserCreatePayload{
					Name:     secondUserName,
					Email:    "castadon@example.com",
					Password: "bobina",
				},
			},
			responseCode: http.StatusOK,
			testLogin:    true,
			testRefresh:  true,
			testName:     "New Email",
		},
		{
			loginPayload: UserLoginPayload{
				Email:    firstUserEmail,
				Password: firstUserPassword,
			},
			payload: UserUpdatePayload{
				UserCreatePayload{
					Name:     "Jonathan Service",
					Email:    "Inqindi@example.com",
					Password: "axel&brie&cindy",
				},
			},
			responseCode: http.StatusOK,
			testLogin:    true,
			testRefresh:  true,
			testName:     "New Email and Password",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			//Login
			loginBody := CreateJSONReader(testCase.loginPayload, t)
			res, err := http.Post(feast_url+"/api/login", "application/json", loginBody)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Expected an OK response code, received: %d", res.StatusCode)
			}

			userLoginResponse := UserLoginResponse{}
			DecodeJSONResponse(&userLoginResponse, res.Body, t)
			res.Body.Close()

			// Update User
			updateBody := CreateJSONReader(testCase.payload, t)
			req, err := http.NewRequest(http.MethodPut, feast_url+"/api/user", updateBody)
			req.Header.Add("Authorization", "Bearer "+userLoginResponse.Token)

			res, err = http.DefaultClient.Do(req)
			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected an %d response code, received: %d", testCase.responseCode, res.StatusCode)
			}

		})
	}

}
