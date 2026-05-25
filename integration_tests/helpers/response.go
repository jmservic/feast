package helpers

import (
	"net/http"
	"testing"
)

type errorResponse struct {
	Error string `'json:"error"`
}

func GetResponseObject[T any](t *testing.T, res *http.Response, expectedCode int) T {
	defer res.Body.Close()
	if res.StatusCode != expectedCode {
		if res.Header.Get("content-type") == "application/json" {
			var errResponse errorResponse
			DecodeJSONResponse(&errResponse, res.Body, t)
			t.Logf("Received response error: %s", errResponse.Error)
		}

		t.Fatalf("Expected status code of %d, got: %d", expectedCode, res.StatusCode)
	}

	var data T
	DecodeJSONResponse(&data, res.Body, t)
	return data
}
