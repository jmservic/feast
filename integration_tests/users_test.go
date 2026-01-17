package integration

import (
	"testing"
)

func TestFailing(t *testing.T) {
	t.Error("Failing just because lol")
}
