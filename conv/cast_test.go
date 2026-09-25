package conv

import (
	"testing"
	"time"

	"github.com/go-sdk/core/testx"
)

func TestToE(t *testing.T) {
	v, err := ToE[int]("123")
	testx.Nil(t, err)
	testx.Equal(t, 123, v)

	v, err = ToE[int](3.99)
	testx.Nil(t, err)
	testx.Equal(t, 3, v)

	b, err := ToE[bool](1)
	testx.Nil(t, err)
	testx.Equal(t, true, b)

	s, err := ToE[string](nil)
	testx.Nil(t, err)
	testx.Empty(t, s)

	v, err = ToE[int]("abc")
	testx.NotNil(t, err)
	testx.Zero(t, v)
}

func TestTo(t *testing.T) {
	testx.Equal(t, 123, To[int]("123"))
	testx.Equal(t, 0, To[int]("abc"))
	testx.Equal(t, "starudream", To[string]("starudream"))
	testx.Equal(t, true, To[bool](1))
	testx.Equal(t, time.Second, To[time.Duration]("1s"))
	testx.Equal(t, 0, To[int](nil))
	testx.Equal(t, "", To[string](nil))
}
