package json

import (
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

// 以下变量是 encoding/json/v2 选项构造函数的别名，可直接传给本包和上游的编解码函数。
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
