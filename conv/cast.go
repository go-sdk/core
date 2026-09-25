package conv

import (
	"github.com/spf13/cast"
)

// ToE 将任意值弱类型转换为目标基础类型，转换失败时返回错误和类型零值。
func ToE[T cast.Basic](v any) (T, error) {
	return cast.ToE[T](v)
}

// To 将任意值弱类型转换为目标基础类型，转换失败时返回类型零值而非错误。
func To[T cast.Basic](v any) T {
	return cast.To[T](v)
}
