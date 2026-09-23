package logging

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestTerminalSupportsColor(t *testing.T) {
	testx.True(t, TerminalSupportsColor("xterm-256color", true, false))
	testx.True(t, TerminalSupportsColor("xterm-256color", false, true))
	testx.False(t, TerminalSupportsColor("xterm-256color", false, false))
	testx.False(t, TerminalSupportsColor("dumb", true, true))
}

func TestConsoleFile(t *testing.T) {
	tests := []struct {
		value string
		want  *os.File
	}{
		{value: "", want: os.Stdout},
		{value: "stdout", want: os.Stdout},
		{value: "other", want: os.Stdout},
		{value: "stderr", want: os.Stderr},
		{value: "STDERR", want: os.Stderr},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			t.Setenv(consoleEnv, test.value)
			testx.Same(t, test.want, consoleFile())
		})
	}
}

func TestSlogUsesZerologConsoleFormat(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewConsoleLogger(buffer, true)
	handler := slog.New(newSlogHandler(&logger))

	handler.Info("loading config file", "file", "config.yaml")
	output := strings.TrimSpace(buffer.String())

	testx.Contains(t, output, " INF ")
	testx.Contains(t, output, "loading config file file=config.yaml")
	testx.NotContains(t, output, "INFO loading config file")
}
