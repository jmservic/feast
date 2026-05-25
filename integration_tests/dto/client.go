package dto

import (
	"testing"
)

type Client struct {
	t        *testing.T
	feastUrl string
}

func NewClient(t *testing.T, feastUrl string) Client {
	return Client{
		t:        t,
		feastUrl: feastUrl,
	}
}
