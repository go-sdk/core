package osx

import (
	"os"
)

// Hostname 返回当前主机名；读取失败时返回空字符串。
func Hostname() string {
	name, _ := os.Hostname()
	return name
}
