package integration

import (
	"bytes"
	"net/http"
	"os"
	"testing"
)

func TestCreateNewUser(t *testing.T) {
	//loadDotEnv()
	feast_url := "http://localhost:" + os.Getenv("PORT")
	body := bytes.NewReader([]byte(`{
		"name": "jonathan",
		"email":  "jon@example.com",
		"password": "very-secret!"
	}`))

	res, err := http.Post(feast_url+"/api/users", "application/json", body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status ok, got: %d", res.StatusCode)
	}

}
