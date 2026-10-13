package dto

import (
	"fmt"
	//	"github.com/google/uuid"
	//	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/constants"
	"github.com/jmservic/feast/integration_tests/helpers"
	//"io"
	"net/http"
)

func (c Client) CreateIngredient(token string, params map[IngredientKeys]any) *http.Response {
	body := helpers.CreateJSONReader(params, c.t)
	/*buffer, err := io.ReadAll(body)
	body.Seek(0, io.SeekStart)
	if err != nil {
		c.t.Fatalf("Error reading requestBody: %v", err)
	}

	jsonBody := string(buffer)
	c.t.Logf("Create Ingredient JSON:\n %s", jsonBody)*/

	req, err := http.NewRequest(http.MethodPost, c.feastUrl+constants.IngredientsPath, body)
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

func (c Client) GetIngredient(token string, id int64) *http.Response {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(c.feastUrl+constants.IngredientsIdPath, id), nil)
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

func (c Client) GetIngredients(token, query string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.feastUrl+constants.IngredientsPath, nil)
	if err != nil {
		c.t.Fatalf("Error occurred when creating the get request: %v", err)
	}
	q := req.URL.Query()
	q.Add("q", query)
	req.URL.RawQuery = q.Encode()
	c.t.Logf("%v", req.URL)

	req.Header.Add("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Unexpected error: %v", err)
	}

	return res
}
