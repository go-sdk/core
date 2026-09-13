package yaml

import (
	"errors"
	"testing"

	"github.com/go-sdk/core/testx"
)

type user struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

type failing struct{}

func (failing) MarshalYAML() (any, error) { return nil, errors.New("marshal failed") }

const raw = "name: john\nage: 18\n"

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
	_, err := Marshal[string](failing{})
	testx.Error(t, err)

	testx.Panics(t, func() { MustMarshal[string](failing{}) })
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
