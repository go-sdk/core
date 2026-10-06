package json

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/go-sdk/core/codec"
	"github.com/go-sdk/core/conv"
	"github.com/go-sdk/core/osx"
)

// MarshalWrite 和 UnmarshalRead 是 encoding/json/v2 对应函数的别名，
// 直接读写 io.Writer 和 io.Reader，避免中间 []byte 载体。
var (
	MarshalWrite  = json.MarshalWrite
	UnmarshalRead = json.UnmarshalRead
)

type Value = jsontext.Value

// 以下变量是 encoding/json/v2 和 jsontext 选项构造函数的别名，可直接传给本包和上游的编解码函数。
var (
	// StringifyNumbers 序列化和反序列化都生效：数字以 JSON 字符串形式编码和解析。
	StringifyNumbers = json.StringifyNumbers
	// Deterministic 仅序列化生效：相同输入始终序列化为相同字节，map 按键排序。
	Deterministic = json.Deterministic
	// FormatNilSliceAsNull 仅序列化生效：nil 切片编码为 null，而非默认的空数组或空字符串。
	FormatNilSliceAsNull = json.FormatNilSliceAsNull
	// FormatNilMapAsNull 仅序列化生效：nil map 编码为 null，而非默认的空对象。
	FormatNilMapAsNull = json.FormatNilMapAsNull
	// OmitZeroStructFields 仅序列化生效：省略零值结构体字段，等价于所有字段附加 omitzero。
	OmitZeroStructFields = json.OmitZeroStructFields
	// MatchCaseInsensitiveNames 序列化和反序列化都生效：JSON 成员名与结构体字段不区分大小写匹配。
	MatchCaseInsensitiveNames = json.MatchCaseInsensitiveNames
	// RejectUnknownMembers 仅反序列化生效：JSON 对象出现未知成员时报错。
	RejectUnknownMembers = json.RejectUnknownMembers

	// WithIndent 仅序列化生效：开启多行输出，指定每层缩进，只允许空格和制表符。
	WithIndent = jsontext.WithIndent
	// WithIndentPrefix 仅序列化生效：开启多行输出，指定行前缀，只允许空格和制表符。
	WithIndentPrefix = jsontext.WithIndentPrefix
	// Multiline 仅序列化生效：控制多行输出，未指定缩进时使用制表符。
	Multiline = jsontext.Multiline
	// EscapeForHTML 仅序列化生效：转义字符串中的 <、> 和 &，用于嵌入 HTML。
	EscapeForHTML = jsontext.EscapeForHTML
	// EscapeForJS 仅序列化生效：转义字符串中的 U+2028 和 U+2029，用于嵌入 JavaScript。
	EscapeForJS = jsontext.EscapeForJS
	// SpaceAfterColon 仅序列化生效：控制冒号后的空格。
	SpaceAfterColon = jsontext.SpaceAfterColon
	// SpaceAfterComma 仅序列化生效：控制逗号后的空格。
	SpaceAfterComma = jsontext.SpaceAfterComma
)

func Marshal[T codec.Data](in any, opts ...json.Options) (T, error) {
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

func MustMarshal[T codec.Data](in any, opts ...json.Options) T {
	v, err := Marshal[T](in, opts...)
	if err != nil {
		osx.Panic(err)
	}
	return v
}

func MarshalIndent[T codec.Data](in any, indent string, opts ...json.Options) (T, error) {
	return Marshal[T](in, append([]json.Options{WithIndent(indent)}, opts...)...)
}

func MustMarshalIndent[T codec.Data](in any, indent string, opts ...json.Options) T {
	return MustMarshal[T](in, append([]json.Options{WithIndent(indent)}, opts...)...)
}

func MarshalOmitZero[T codec.Data](in any, opts ...json.Options) (T, error) {
	return Marshal[T](in, append([]json.Options{OmitZeroStructFields(true)}, opts...)...)
}

func MustMarshalOmitZero[T codec.Data](in any, opts ...json.Options) T {
	return MustMarshal[T](in, append([]json.Options{OmitZeroStructFields(true)}, opts...)...)
}

func Unmarshal[T codec.Data](in T, out any, opts ...json.Options) error {
	var bs []byte
	switch v := any(in).(type) {
	case string:
		bs = conv.StringToBytes(v)
	case []byte:
		bs = v
	}
	return json.Unmarshal(bs, out, opts...)
}

func MustUnmarshal[V any, T codec.Data](in T, opts ...json.Options) (v V) {
	if err := Unmarshal(in, &v, opts...); err != nil {
		osx.Panic(err)
	}
	return v
}
