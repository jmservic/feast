package dto

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
	"strings"
	"testing"
)

func (c Client) CreateUser(name, email, password string) *http.Response {
	payload := UserCreatePayload{
		Name:     name,
		Email:    email,
		Password: password,
	}

	body := helpers.CreateJSONReader(payload, c.t)

	res, err := http.Post(c.feastUrl+constants.UsersPath, "application/json", body)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}

func ValidateUserCreateResponse(t *testing.T, userCreateResponse UserCreateResponse, name, email string) {
	if userCreateResponse.Name != name {
		t.Fatalf("Expected %s, but got %s for the name", name, userCreateResponse.Name)
	}
	if userCreateResponse.Email != strings.ToLower(email) {
		t.Fatalf("Expected %s, but got %s for the email", email, userCreateResponse.Email)
	}
	if userCreateResponse.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the user id")
	}

}

func (c Client) LoginUser(email, password string) *http.Response {
	payload := UserLoginPayload{
		Email:    email,
		Password: password,
	}
	body := helpers.CreateJSONReader(payload, c.t)
	res, err := http.Post(c.feastUrl+constants.LoginPath, "application/json", body)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}
	return res
}

func (c Client) RefreshUser(refreshToken string) *http.Response {
	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.RefreshPath, nil)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the refresh request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+refreshToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res

}

func UpdateUser(t *testing.T, feastUrl, token, name, email, password string) *http.Response {
	payload := UserUpdatePayload{
		UserCreatePayload: UserCreatePayload{
			Name:     name,
			Email:    email,
			Password: password,
		},
	}
	updateBody := helpers.CreateJSONReader(payload, t)
	req, err := http.NewRequest(http.MethodPut, feastUrl+constants.UsersPath, updateBody)
	if err != nil {
		t.Fatalf("Error occurred when creating the update request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return res
}

func DeleteUser(t *testing.T, feastUrl, token string) *http.Response {
	req, err := http.NewRequest(http.MethodDelete, feastUrl+constants.UsersPath, nil)
	if err != nil {
		t.Fatalf("error occurred when creating the delete request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return res
}
