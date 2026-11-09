package dto

import (
	//	"fmt"
	//	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	//"io"
	"net/http"
)

func (c Client) CreateFoodItem(token string, params map[FoodItemKeys]any) *http.Response {
	body := helpers.CreateJSONReader(params, c.t)

	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.FoodLogsPath, body)
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
