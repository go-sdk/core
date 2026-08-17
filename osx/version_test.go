package osx

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestGetVersion(t *testing.T) {
	got := GetVersion()

	testx.NotEmpty(t, got.Version)
	testx.Equal(t, runtime.Version(), got.GoVersion)
	testx.Equal(t, runtime.Compiler, got.Compiler)
	testx.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, got.Platform)
	testx.Equal(t, got, GetVersion())
}

func TestVersionString(t *testing.T) {
	want := Version{
		Type:      "git",
		Version:   "v1.2.3",
		Revision:  "abc123",
		Time:      "2026-08-17T12:00:00+08:00",
		Dirty:     true,
		GoVersion: "go1.26",
		Compiler:  "gc",
		Platform:  "darwin/arm64",
	}

	text := want.String()
	testx.Contains(t, text, "\n")

	var got Version
	testx.NoError(t, json.Unmarshal([]byte(text), &got))
	testx.Equal(t, want, got)
}
