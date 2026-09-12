package osx

import (
	"io"
	"os"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestPanic(t *testing.T) {
	original := os.Stderr
	r, w, err := os.Pipe()
	testx.NoError(t, err)
	os.Stderr = w

	var value any
	func() {
		defer func() { value = recover() }()
		Panic("boom")
	}()

	testx.NoError(t, w.Close())
	os.Stderr = original

	output, err := io.ReadAll(r)
	testx.NoError(t, err)
	testx.NoError(t, r.Close())

	testx.Equal(t, "boom", value)

	s := string(output)
	testx.Contains(t, s, "[PANIC]")
	testx.Contains(t, s, "boom")
	testx.Contains(t, s, "TestPanic")
	testx.Contains(t, s, "panic_test.go")
	testx.NotContains(t, s, "osx.Panic\n")
}

func TestPanicf(t *testing.T) {
	original := os.Stderr
	r, w, err := os.Pipe()
	testx.NoError(t, err)
	os.Stderr = w

	var value any
	func() {
		defer func() { value = recover() }()
		Panicf("boom: %d items", 3)
	}()

	testx.NoError(t, w.Close())
	os.Stderr = original

	output, err := io.ReadAll(r)
	testx.NoError(t, err)
	testx.NoError(t, r.Close())

	testx.Equal(t, "boom: 3 items", value)

	s := string(output)
	testx.Contains(t, s, "[PANIC]")
	testx.Contains(t, s, "boom: 3 items")
	testx.Contains(t, s, "TestPanicf")
	testx.Contains(t, s, "panic_test.go")
	testx.NotContains(t, s, "osx.Panicf\n")
}
