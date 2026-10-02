package yaml

import (
	"errors"
	"io"
	"testing"

	stdyaml "go.yaml.in/yaml/v3"

	"github.com/go-sdk/core/testx"
)

type user struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

type failing struct{}

func (failing) MarshalYAML() (any, error) { return nil, errors.New("marshal failed") }

type eofUnmarshaler struct {
	calls int
}

func (value *eofUnmarshaler) UnmarshalYAML(*stdyaml.Node) error {
	value.calls++
	return io.EOF
}

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

func TestMarshalOptions(t *testing.T) {
	in := map[string][]string{"items": {"john", "jane"}}
	tests := []struct {
		name string
		opts []Options
		want string
	}{
		{"default", nil, "items:\n    - john\n    - jane\n"},
		{"zero-indent", []Options{WithIndent(0)}, "items:\n    - john\n    - jane\n"},
		{"indent", []Options{WithIndent(2)}, "items:\n  - john\n  - jane\n"},
		{"compact", []Options{WithIndent(2), WithCompactSeqIndent(true)}, "items:\n- john\n- jane\n"},
		{"compact-default-indent", []Options{WithCompactSeqIndent(true)}, "items:\n  - john\n  - jane\n"},
		{"reset-compact", []Options{WithIndent(2), WithCompactSeqIndent(true), WithCompactSeqIndent(false)}, "items:\n  - john\n  - jane\n"},
		{"later-indent-wins", []Options{WithIndent(4), WithIndent(2)}, "items:\n  - john\n  - jane\n"},
		{"known-fields-ignored", []Options{WithKnownFields(true)}, "items:\n    - john\n    - jane\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, err := Marshal[string](in, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, text)
			data, err := Marshal[[]byte](in, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, string(data))
			testx.Equal(t, test.want, MustMarshal[string](in, test.opts...))
		})
	}
}

func TestMarshalOptionsError(t *testing.T) {
	text, err := Marshal[string](failing{}, WithIndent(2))
	testx.Error(t, err)
	testx.Equal(t, "", text)
	data, err := Marshal[[]byte](failing{}, WithIndent(2))
	testx.Error(t, err)
	testx.Nil(t, data)
	testx.Panics(t, func() { MustMarshal[string](failing{}, WithIndent(2)) })
	testx.Panics(t, func() { _, _ = Marshal[string](want, WithIndent(-1)) })
}

func TestKnownFields(t *testing.T) {
	input := raw + "extra: value\n"
	var out user
	testx.Error(t, Unmarshal(input, &out, WithKnownFields(true)))
	testx.Error(t, Unmarshal([]byte(input), &out, WithKnownFields(true)))
	testx.Panics(t, func() { MustUnmarshal[user](input, WithKnownFields(true)) })
	testx.NoError(t, Unmarshal(input, &out, WithKnownFields(true), WithKnownFields(false)))
	testx.Equal(t, want, out)
	testx.Equal(t, want, MustUnmarshal[user](raw, WithKnownFields(true)))
	testx.Equal(t, want, MustUnmarshal[user, []byte]([]byte(raw), WithKnownFields(true)))
	var mapping map[string]any
	testx.NoError(t, Unmarshal(input, &mapping, WithKnownFields(true)))
	testx.Equal(t, "value", mapping["extra"])
}

func TestUnmarshalOptions(t *testing.T) {
	for _, input := range []string{"", "# comment\n", "---\n", raw, raw + "---\nname: jane\n"} {
		t.Run(input, func(t *testing.T) {
			baseline := user{Name: "existing", Age: 42}
			testx.NoError(t, Unmarshal(input, &baseline))
			out := user{Name: "existing", Age: 42}
			testx.NoError(t, Unmarshal(input, &out, WithKnownFields(true), WithIndent(-1), WithCompactSeqIndent(true)))
			testx.Equal(t, baseline, out)
			out = user{Name: "existing", Age: 42}
			testx.NoError(t, Unmarshal([]byte(input), &out, WithKnownFields(true)))
			testx.Equal(t, baseline, out)
		})
	}
	var out user
	testx.Error(t, Unmarshal("name: [", &out, WithKnownFields(true)))
}

func TestUnmarshalCustomEOF(t *testing.T) {
	tests := []struct {
		name string
		opts []Options
	}{
		{"default", nil},
		{"known-fields", []Options{WithKnownFields(true)}},
		{"relaxed-fields", []Options{WithKnownFields(false)}},
		{"encoding-options", []Options{WithIndent(2), WithCompactSeqIndent(true)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var fromString eofUnmarshaler
			testx.ErrorIs(t, Unmarshal(raw, &fromString, test.opts...), io.EOF)
			testx.Equal(t, 1, fromString.calls)
			var fromBytes eofUnmarshaler
			testx.ErrorIs(t, Unmarshal([]byte(raw), &fromBytes, test.opts...), io.EOF)
			testx.Equal(t, 1, fromBytes.calls)
			testx.Panics(t, func() { MustUnmarshal[eofUnmarshaler](raw, test.opts...) })
			testx.Panics(t, func() { MustUnmarshal[eofUnmarshaler, []byte]([]byte(raw), test.opts...) })
		})
	}
	var nested struct {
		Value eofUnmarshaler `yaml:"value"`
	}
	testx.ErrorIs(t, Unmarshal("value: john\n", &nested, WithKnownFields(true)), io.EOF)
	testx.Equal(t, 1, nested.Value.calls)
}

func TestUnmarshalEmptyCustomEOF(t *testing.T) {
	for _, input := range []string{"", " \n", "# comment\n"} {
		t.Run(input, func(t *testing.T) {
			var fromString eofUnmarshaler
			testx.NoError(t, Unmarshal(input, &fromString, WithKnownFields(true)))
			testx.Equal(t, 0, fromString.calls)
			var fromBytes eofUnmarshaler
			testx.NoError(t, Unmarshal([]byte(input), &fromBytes, WithKnownFields(true)))
			testx.Equal(t, 0, fromBytes.calls)
		})
	}
}
