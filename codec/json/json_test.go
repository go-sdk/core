package json

import (
	"testing"

	"github.com/go-sdk/core/testx"
)

type user struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

const raw = `{"name":"john","age":18}`

var want = user{Name: "john", Age: 18}

func TestMarshal(t *testing.T) {
	s, err := Marshal[string](want)
	testx.NoError(t, err)
	testx.Equal(t, raw, s)

	bs, err := Marshal[[]byte](want)
	testx.NoError(t, err)
	testx.Equal(t, raw, string(bs))
}

func TestMustMarshal(t *testing.T) {
	testx.Equal(t, raw, MustMarshal[string](want))
	testx.Equal(t, raw, string(MustMarshal[[]byte](want)))
}

func TestMarshalError(t *testing.T) {
	_, err := Marshal[string](make(chan int))
	testx.Error(t, err)

	testx.Panics(t, func() { MustMarshal[string](make(chan int)) })
}

func TestUnmarshal(t *testing.T) {
	var fromString user
	testx.NoError(t, Unmarshal(raw, &fromString))
	testx.Equal(t, want, fromString)

	var fromBytes user
	testx.NoError(t, Unmarshal([]byte(raw), &fromBytes))
	testx.Equal(t, want, fromBytes)
}

func TestMustUnmarshal(t *testing.T) {
	testx.Equal(t, want, MustUnmarshal[user](raw))
	testx.Equal(t, want, MustUnmarshal[user, []byte]([]byte(raw)))
}
