package integration

import (
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"testing"
)

// TO-DO: Add failing test cases like a bad name, email, or password
func TestCreateNewUser(t *testing.T) {
	loadDotEnv()
	name := "jonathan"
	email := "jon@example.com"
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
	if sut.Email != email {
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Fatal("I'm just going to fail because")
		})

	}
}
