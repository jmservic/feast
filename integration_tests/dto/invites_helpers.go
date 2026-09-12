package dto

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	//"github.com/jmservic/feast/integration_tests/helpers"
	"encoding/json"
	"net/http"
	"net/url"
)

func (c Client) GetInvites(token string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.feastUrl+constants.InvitesPath, nil)
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

func (c Client) InviteUserToHousehold(token string, userId, householdId uuid.UUID, memberId *uuid.UUID) *http.Response {
	payload := InvitePayload{
		UserId:            userId,
		HouseholdId:       householdId,
		HouseholdMemberId: memberId,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		c.t.Fatalf("Error marshaling the InvitePayload: %v", err)
	}

	body := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.InvitesPath, body)
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

func (c Client) HandleInvite(token string, household_id uuid.UUID, accept bool) *http.Response {
	formPayload := url.Values{}
	formPayload.Add("household_id", household_id.String())

	if accept {
		formPayload.Add("action", "accept")
	} else {
		formPayload.Add("action", "decline")
	}

	data := []byte(formPayload.Encode())
	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPut, c.feastUrl+constants.InvitesPath, dataReader)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}
