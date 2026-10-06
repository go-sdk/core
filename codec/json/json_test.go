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

func TestMarshalIndent(t *testing.T) {
	tests := []struct {
		name   string
		indent string
		opts   []stdjson.Options
		want   string
	}{
		{
			name:   "spaces",
			indent: "  ",
			want:   "{\n  \"name\": \"john\",\n  \"age\": 18\n}",
		},
		{
			name:   "tab",
			indent: "\t",
			want:   "{\n\t\"name\": \"john\",\n\t\"age\": 18\n}",
		},
		{
			name:   "empty-indent",
			indent: "",
			want:   "{\n\"name\": \"john\",\n\"age\": 18\n}",
		},
		{
			name:   "override-indent",
			indent: "  ",
			opts:   []stdjson.Options{WithIndent("\t")},
			want:   "{\n\t\"name\": \"john\",\n\t\"age\": 18\n}",
		},
		{
			name:   "single-line",
			indent: "  ",
			opts:   []stdjson.Options{Multiline(false), SpaceAfterColon(false)},
			want:   raw,
		},
		{
			name:   "stringify-numbers",
			indent: "  ",
			opts:   []stdjson.Options{StringifyNumbers(true)},
			want:   "{\n  \"name\": \"john\",\n  \"age\": \"18\"\n}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, err := MarshalIndent[string](want, test.indent, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, text)
			data, err := MarshalIndent[[]byte](want, test.indent, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, string(data))
			testx.Equal(t, test.want, MustMarshalIndent[string](want, test.indent, test.opts...))
			testx.Equal(t, test.want, string(MustMarshalIndent[[]byte](want, test.indent, test.opts...)))
		})
	}
}

func TestMarshalIndentError(t *testing.T) {
	text, err := MarshalIndent[string](make(chan int), "  ")
	testx.Error(t, err)
	testx.Empty(t, text)
	data, err := MarshalIndent[[]byte](make(chan int), "  ")
	testx.Error(t, err)
	testx.Nil(t, data)
	testx.Panics(t, func() { MustMarshalIndent[string](make(chan int), "  ") })
	testx.Panics(t, func() { MustMarshalIndent[[]byte](make(chan int), "  ") })
}

func TestMarshalOmitZero(t *testing.T) {
	tests := []struct {
		name string
		in   any
		opts []stdjson.Options
		want string
	}{
		{
			name: "non-zero",
			in:   want,
			want: raw,
		},
		{
			name: "zero-field",
			in:   user{Name: "john"},
			want: `{"name":"john"}`,
		},
		{
			name: "all-zero",
			in:   user{},
			want: `{}`,
		},
		{
			name: "disable-omit-zero",
			in:   user{},
			opts: []stdjson.Options{OmitZeroStructFields(false)},
			want: `{"name":"","age":0}`,
		},
		{
			name: "indent",
			in:   user{Name: "john"},
			opts: []stdjson.Options{WithIndent("  ")},
			want: "{\n  \"name\": \"john\"\n}",
		},
		{
			name: "nil-slice",
			in: struct {
				Items []string `json:"items"`
			}{},
			want: `{}`,
		},
		{
			name: "empty-slice",
			in: struct {
				Items []string `json:"items"`
			}{Items: []string{}},
			want: `{"items":[]}`,
		},
		{
			name: "map-zero-value",
			in:   map[string]int{"age": 0},
			want: `{"age":0}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, err := MarshalOmitZero[string](test.in, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, text)
			data, err := MarshalOmitZero[[]byte](test.in, test.opts...)
			testx.NoError(t, err)
			testx.Equal(t, test.want, string(data))
			testx.Equal(t, test.want, MustMarshalOmitZero[string](test.in, test.opts...))
			testx.Equal(t, test.want, string(MustMarshalOmitZero[[]byte](test.in, test.opts...)))
		})
	}
}

func TestMarshalOmitZeroError(t *testing.T) {
	text, err := MarshalOmitZero[string](make(chan int))
	testx.Error(t, err)
	testx.Empty(t, text)
	data, err := MarshalOmitZero[[]byte](make(chan int))
	testx.Error(t, err)
	testx.Nil(t, data)
	testx.Panics(t, func() { MustMarshalOmitZero[string](make(chan int)) })
	testx.Panics(t, func() { MustMarshalOmitZero[[]byte](make(chan int)) })
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
