package json

import (
	stdjson "encoding/json/v2"
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

func TestJSONTextOptions(t *testing.T) {
	tests := []struct {
		name string
		in   any
		opts []stdjson.Options
		want string
	}{
		{
			name: "indent",
			in:   want,
			opts: []stdjson.Options{WithIndent("  ")},
			want: "{\n  \"name\": \"john\",\n  \"age\": 18\n}",
		},
		{
			name: "prefix",
			in:   want,
			opts: []stdjson.Options{WithIndent("  "), WithIndentPrefix("\t")},
			want: "{\n\t  \"name\": \"john\",\n\t  \"age\": 18\n\t}",
		},
		{
			name: "multiline",
			in:   want,
			opts: []stdjson.Options{Multiline(true)},
			want: "{\n\t\"name\": \"john\",\n\t\"age\": 18\n}",
		},
		{
			name: "single-line",
			in:   want,
			opts: []stdjson.Options{WithIndent("  "), Multiline(false), SpaceAfterColon(false)},
			want: raw,
		},
		{
			name: "html",
			in:   "<>&",
			opts: []stdjson.Options{EscapeForHTML(true)},
			want: `"\u003c\u003e\u0026"`,
		},
		{
			name: "javascript",
			in:   "\u2028\u2029",
			opts: []stdjson.Options{EscapeForJS(true)},
			want: `"\u2028\u2029"`,
		},
		{
			name: "spaces",
			in:   want,
			opts: []stdjson.Options{SpaceAfterColon(true), SpaceAfterComma(true)},
			want: `{"name": "john", "age": 18}`,
		},
		{
			name: "later-option-wins",
			in:   want,
			opts: []stdjson.Options{SpaceAfterColon(true), SpaceAfterColon(false)},
			want: raw,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, err := Marshal[string](test.in, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, text)
			data, err := Marshal[[]byte](test.in, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, string(data))
			testx.Equal(t, test.want, MustMarshal[string](test.in, test.opts...))
		})
	}
}

func TestJSONTextOptionsIgnoredOnUnmarshal(t *testing.T) {
	opts := []stdjson.Options{WithIndent("  "), WithIndentPrefix("\t"), Multiline(true), EscapeForHTML(true), EscapeForJS(true), SpaceAfterColon(true), SpaceAfterComma(true)}
	var out user
	testx.NoError(t, Unmarshal(raw, &out, opts...))
	testx.Equal(t, want, out)
}
