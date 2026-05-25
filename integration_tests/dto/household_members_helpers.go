package dto

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
)

func (c Client) CreateHouseholdMember(token, name string, householdId uuid.UUID, userId *uuid.UUID) *http.Response {
	payload := HouseholdMemberPayload{
		Name:   name,
		UserId: userId,
	}

	body := helpers.CreateJSONReader(payload, c.t)

	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.HouseholdsPath+"/"+householdId.String()+constants.HouseMembersPath, body)
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

func (c Client) GetHouseholdMembers(token string, householdId uuid.UUID) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.feastUrl+constants.HouseholdsPath+"/"+householdId.String()+constants.HouseMembersPath, nil)
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
