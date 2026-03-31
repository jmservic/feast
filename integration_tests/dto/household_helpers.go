package dto

import (
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
	"testing"
)

func CreateHousehold(t *testing.T, feastUrl, token, name string) *http.Response {
	payload := HouseholdCreatePayload{
		Name: name,
	}

	body := helpers.CreateJSONReader(payload, t)
	req, err := http.NewRequest(http.MethodPost, feastUrl+constants.HouseholdsPath, body)
	if err != nil {
		t.Fatalf("Error occurred when creating the create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	return res
}
