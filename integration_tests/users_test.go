package integration

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// TO-DO: Add failing test cases like a bad name, email, or password
func TestCreateNewUser(t *testing.T) {
	loadDotEnv()
	name := "jonathan"
	email := "jon@example.com"
	password := "very-secret!"

	feast_url := "http://localhost:" + os.Getenv("PORT")
	t.Cleanup(func() { resetDatabase(feast_url) })

	payload := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     name,
		Email:    email,
		Password: password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Error marshaling the payload: %s", err)
	}
	body := bytes.NewReader(payloadBytes)

	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	sut := struct {
		Id        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{}

	if err := decoder.Decode(&sut); err != nil {
		t.Fatalf("Unexpected error decoding the json response: %v", err)
	}

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

	feast_url := "http://localhost:" + os.Getenv("PORT")
	t.Cleanup(func() { resetDatabase(feast_url) })

	payload := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     name,
		Email:    email,
		Password: password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Error marshaling the payload: %s", err)
	}
	body := bytes.NewReader(payloadBytes)

	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	sut := struct {
		Id        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{}

	if err := decoder.Decode(&sut); err != nil {
		t.Fatalf("Unexpected error decoding the json response: %v", err)
	}

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
