package yaml

import (
	"go.yaml.in/yaml/v3"

	"github.com/go-sdk/core/codec"
	"github.com/go-sdk/core/conv"
	"github.com/go-sdk/core/osx"
)

func Marshal[T codec.Data](in any) (T, error) {
	bs, err := yaml.Marshal(in)
	if err != nil {
		var zero T
		return zero, err
	}
	var out T
	switch v := any(&out).(type) {
	case *string:
		*v = conv.BytesToString(bs)
	case *[]byte:
		*v = bs
	}
	return out, nil
}

func MustMarshal[T codec.Data](in any) T {
	v, err := Marshal[T](in)
	if err != nil {
		osx.Panic(err)
	}
	return v
}

func Unmarshal[T codec.Data](in T, out any) error {
	var bs []byte
	switch v := any(in).(type) {
	case string:
		bs = conv.StringToBytes(v)
	case []byte:
		bs = v
	}
	return yaml.Unmarshal(bs, out)
}

func MustUnmarshal[V any, T codec.Data](in T) (v V) {
	if err := Unmarshal(in, &v); err != nil {
		osx.Panic(err)
	}
	return v
}
