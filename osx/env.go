package osx

import (
	"os"

	"github.com/spf13/cast"
)

// GetEnv 按 names 的先后顺序读取并转换第一个已设置的环境变量。
// 环境变量即使为空或转换失败也会立即返回转换结果，不再尝试后续名称；
// 所有环境变量都未设置时返回 value。
func GetEnv[T cast.Basic](value T, names ...string) T {
	for _, name := range names {
		if v, ok := os.LookupEnv(name); ok {
			t, _ := cast.ToE[T](v)
			return t
		}
	}
	return value
}
