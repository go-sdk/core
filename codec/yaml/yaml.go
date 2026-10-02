package yaml

import (
	"bytes"
	"io"

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

// Options 配置编解码行为，后面的同名选项覆盖前面的选项；不适用于当前方向的选项被忽略。
type Options func(*options)

type options struct {
	indent           int
	compactSeqIndent bool
	knownFields      bool
}

// WithIndent 仅序列化生效：设置缩进空格数，0 使用上游默认值，负数在序列化时触发 panic。
func WithIndent(spaces int) Options {
	return func(opts *options) { opts.indent = spaces }
}

// WithCompactSeqIndent 仅序列化生效：控制序列标记 "- " 是否计入缩进，默认关闭。
func WithCompactSeqIndent(enabled bool) Options {
	return func(opts *options) { opts.compactSeqIndent = enabled }
}

// WithKnownFields 仅反序列化生效：检查映射键是否对应目标结构体字段，默认关闭，不限制 map 的键。
func WithKnownFields(enabled bool) Options {
	return func(opts *options) { opts.knownFields = enabled }
}

func Marshal[T codec.Data](in any, opts ...Options) (T, error) {
	var bs []byte
	var err error
	if len(opts) == 0 {
		bs, err = yaml.Marshal(in)
	} else {
		var config options
		for _, opt := range opts {
			opt(&config)
		}
		var buffer bytes.Buffer
		encoder := yaml.NewEncoder(&buffer)
		encoder.SetIndent(config.indent)
		if config.compactSeqIndent {
			encoder.CompactSeqIndent()
		}
		err = encoder.Encode(in)
		closeErr := encoder.Close()
		if err == nil {
			err = closeErr
		}
		bs = buffer.Bytes()
	}
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

func MustMarshal[T codec.Data](in any, opts ...Options) T {
	v, err := Marshal[T](in, opts...)
	if err != nil {
		osx.Panic(err)
	}
	return v
}

func Unmarshal[T codec.Data](in T, out any, opts ...Options) error {
	var bs []byte
	switch v := any(in).(type) {
	case string:
		bs = conv.StringToBytes(v)
	case []byte:
		bs = v
	}
	if len(opts) == 0 {
		return yaml.Unmarshal(bs, out)
	}
	var config options
	for _, opt := range opts {
		opt(&config)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(bs))
	decoder.KnownFields(config.knownFields)
	err := decoder.Decode(out)
	if err == io.EOF {
		var node yaml.Node
		if parseErr := yaml.Unmarshal(bs, &node); parseErr == nil && node.Kind == 0 {
			return nil
		}
	}
	return err
}

func MustUnmarshal[V any, T codec.Data](in T, opts ...Options) (v V) {
	if err := Unmarshal(in, &v, opts...); err != nil {
		osx.Panic(err)
	}
	return v
}
