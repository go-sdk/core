package osx

import (
	"os"
	"strings"
)

// IsDebug 判断程序是否处于调试模式。
// 命令行包含以 "-test." 开头的参数，或 DEBUG 环境变量为
// 1、t、true、y、yes、on 中的任意值时返回 true。
func IsDebug() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	switch strings.ToLower(os.Getenv("DEBUG")) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	}
	return false
}
