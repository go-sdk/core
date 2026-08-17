package osx

import (
	"os"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestHostname(t *testing.T) {
	want, err := os.Hostname()
	testx.NoError(t, err)
	testx.Equal(t, want, Hostname())
}
