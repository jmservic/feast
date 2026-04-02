package helpers

import (
	"net/http"
	"testing"
)

func GetResponseObject[T any](t *testing.T, res *http.Response, expectedCode int) T {
	defer res.Body.Close()
	if res.StatusCode != expectedCode {
		t.Fatalf("Expected status code of %d, got: %d", expectedCode, res.StatusCode)
	}

	var data T
	DecodeJSONResponse(&data, res.Body, t)
	return data
}
