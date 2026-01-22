package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func CreateJSONReader[T any](payload T, t *testing.T) *bytes.Reader {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Error marshaling the payload: %s", err)
	}
	reader := bytes.NewReader(payloadBytes)
	return reader
}

func DecodeJSONResponse[T any](responseStruct *T, body io.ReadCloser, t *testing.T) {
	decoder := json.NewDecoder(body)

	if err := decoder.Decode(responseStruct); err != nil {
		t.Fatalf("Unexpected error decoding the json response: %v", err)
	}
}
