package dto

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	//"github.com/jmservic/feast/integration_tests/helpers"
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

func (c Client) HandleInvite(token string, household_id uuid.UUID, accept bool) *http.Response {
	//household_id, action
	c.t.Logf("%v", household_id)
	formPayload := url.Values{}
	formPayload.Add("household_id", household_id.String())
	//payload := InviteHandlePayload{
	//	HouseholdId: household_id,
	//}

	if accept {
		//	payload.Action = "accept"
		formPayload.Add("action", "accept")
	} else {
		//payload.Action = "decline"
		formPayload.Add("action", "decline")
	}

	//data := helpers.CreateJSONReader(payload, c.t)

	data := []byte(formPayload.Encode())
	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.InvitesPath, dataReader)
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
