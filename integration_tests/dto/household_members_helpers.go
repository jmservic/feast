package dto

import (
	"fmt"
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
	path := fmt.Sprintf(constants.HouseholdIdMembersPath, householdId.String())
	req, err := http.NewRequest(http.MethodPost, c.feastUrl+path, body)
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
	path := fmt.Sprintf(constants.HouseholdIdMembersPath, householdId.String())

	req, err := http.NewRequest(http.MethodGet, c.feastUrl+path, nil)
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

func (c Client) GetHouseholdMember(token string, memberId uuid.UUID) *http.Response {
	path := constants.HouseholdMembersPath + "/" + memberId.String()

	req, err := http.NewRequest(http.MethodGet, c.feastUrl+path, nil)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the get request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}

func (c Client) UpdateHouseholdMember(token string, memberId, newUserId uuid.UUID, newName string, newRole int) *http.Response {
	payload := HouseholdMemberUpdatePayload{
		Name:   newName,
		UserId: newUserId,
		Role:   newRole,
	}

	body := helpers.CreateJSONReader(payload, c.t)
	path := constants.HouseholdMembersPath + "/" + memberId.String()

	req, err := http.NewRequest(http.MethodPut, c.feastUrl+path, body)
	if err != nil {
		c.t.Fatalf("Error occurred whencreating the update request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}

func (c Client) DeleteHouseholdMember(token string, memberId uuid.UUID) *http.Response {
	path := constants.HouseholdMembersPath + "/" + memberId.String()
	req, err := http.NewRequest(http.MethodDelete, c.feastUrl+path, nil)
	if err != nil {
		c.t.Fatalf("Error occurred whencreating the update request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}
