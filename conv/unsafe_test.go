package conv

import (
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestStringToBytes(t *testing.T) {
	testx.Nil(t, StringToBytes(""))
	testx.Equal(t, []byte("starudream"), StringToBytes("starudream"))
}

func TestBytesToString(t *testing.T) {
	testx.Equal(t, "", BytesToString(nil))
	testx.Equal(t, "", BytesToString([]byte{}))
	testx.Equal(t, "starudream", BytesToString([]byte("starudream")))
}
