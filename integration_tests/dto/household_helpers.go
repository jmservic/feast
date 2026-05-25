package dto

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
	"testing"
)

func (c Client) CreateHousehold(token, name string) *http.Response {
	payload := HouseholdPayload{
		Name: name,
	}

	body := helpers.CreateJSONReader(payload, c.t)
	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.HouseholdsPath, body)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}

func ValidateHouseholdCreateResponse(t *testing.T, householdCreateResponse HouseholdResponse, name string) {
	if householdCreateResponse.Name != name {
		t.Fatalf("Expected %s, but got %s for the name", name, householdCreateResponse.Name)
	}
	if householdCreateResponse.Id == uuid.Nil {
		t.Fatal("Got a Nil UUID for the household id")
	}
}

func (c Client) UpdateHousehold(householdId, name, token string) *http.Response {
	payload := HouseholdPayload{
		Name: name,
	}
	body := helpers.CreateJSONReader(payload, c.t)
	req, err := http.NewRequest(http.MethodPut, c.feastUrl+constants.HouseholdsPath+"/"+householdId, body)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}

func (c Client) GetHousehold(householdId, token string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.feastUrl+constants.HouseholdsPath+"/"+householdId, nil)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the create request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}

func (c Client) DeleteHousehold(householdId, token string) *http.Response {
	req, err := http.NewRequest(http.MethodDelete, c.feastUrl+constants.HouseholdsPath+"/"+householdId, nil)
	if err != nil {
		c.t.Fatalf("Error occured when creating the delete request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}
