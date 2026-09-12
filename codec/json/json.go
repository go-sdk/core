package json

import (
	"encoding/json/v2"

	"github.com/go-sdk/core/conv"
	"github.com/go-sdk/core/osx"
)

// Data 约束 JSON 序列化结果的载体类型。
type Data interface {
	string | []byte
}

func Marshal[T Data](in any, opts ...json.Options) (T, error) {
	bs, err := json.Marshal(in, opts...)
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

func MustMarshal[T Data](in any, opts ...json.Options) T {
	v, err := Marshal[T](in, opts...)
	if err != nil {
		osx.Panic(err)
	}
	return v
}

func Unmarshal[T Data](in T, out any, opts ...json.Options) error {
	var bs []byte
	switch v := any(in).(type) {
	case string:
		bs = conv.StringToBytes(v)
	case []byte:
		bs = v
	}
	return json.Unmarshal(bs, out, opts...)
}

func MustUnmarshal[V any, T Data](in T, opts ...json.Options) (v V) {
	if err := Unmarshal(in, &v, opts...); err != nil {
		osx.Panic(err)
	}
	return v
}
