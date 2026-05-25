package dto

import (
	"github.com/jmservic/feast/integration_tests/constants"
	"net/http"
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
