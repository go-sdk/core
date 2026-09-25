package conv

import (
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestPtrValue(t *testing.T) {
	v := 123
	testx.Equal(t, 123, PtrValue(&v))
	testx.Equal(t, 0, PtrValue[int](nil))
	testx.Equal(t, "", PtrValue[string](nil))
}
