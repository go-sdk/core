package yaml

import (
	"go.yaml.in/yaml/v3"

	"github.com/go-sdk/core/codec"
	"github.com/go-sdk/core/conv"
	"github.com/go-sdk/core/osx"
)

// NewEncoder 和 NewDecoder 是 go.yaml.in/yaml/v3 对应函数的别名，用于面向 io.Writer
// 和 io.Reader 的流式多文档编解码；Encoder 使用后必须调用 Close，否则剩余数据不会写入 Writer。
var (
	NewEncoder = yaml.NewEncoder
	NewDecoder = yaml.NewDecoder
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
