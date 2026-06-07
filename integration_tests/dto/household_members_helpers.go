package dto

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
)

func (c Client) ValidateHouseholdMemberResponse(expected, actual HouseholdMemberResponse) {
	if actual.Name != expected.Name {
		c.t.Errorf("Name mismatch - expected: %v, actual: %v", expected.Name, actual.Name)
	}

	if actual.Role != expected.Role {
		c.t.Errorf("Role mismatch - expected: %v, actual: %v", expected.Role, actual.Role)
	}

	if actual.HouseholdId != expected.HouseholdId {
		c.t.Errorf("Household Id mismatch - expected: %v, actual: %v", expected.HouseholdId, actual.HouseholdId)
	}

	if actual.UserId != nil && expected.UserId != nil {
		if actual.UserId != expected.UserId {
			c.t.Errorf("UserId mismatch - expected: %v, actual: %v",
				expected.UserId.String(),
				expected.UserId.String())
		}
	} else if actual.UserId == nil && expected.UserId != nil {
		c.t.Errorf("UserId mismatch - expected: %v, actual: nil", expected.UserId.String())
	} else if expected.UserId == nil && actual.UserId != nil {
		c.t.Errorf("UserId mismatch - expected: nil, actual: %v", actual.UserId.String())
	}

	if c.t.Failed() {
		c.t.FailNow()
	}
}

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

func (c Client) GetHouseholdMember(token string, householdId, memberId uuid.UUID) *http.Response {
	req, err := http.NewRequest(http.MethodGet,
		c.feastUrl+constants.HouseholdsPath+"/"+householdId.String()+constants.HouseMembersPath+"/"+memberId.String(), nil)
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
