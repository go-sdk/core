package osx

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// WorkDir 返回当前工作目录；读取失败时返回空字符串。
func WorkDir() string {
	dir, _ := os.Getwd()
	return dir
}

var (
	exe     string
	exeOnce sync.Once
)

func init() {
	exeOnce.Do(func() {
		exe, _ = os.Executable()
	})
}

// ExeFull 返回包初始化时读取并缓存的可执行文件完整路径。
func ExeFull() string {
	return exe
}

// ExeDir 返回可执行文件所在目录。
func ExeDir() string {
	return filepath.Dir(exe)
}

// ExeName 返回不包含目录和扩展名的可执行文件名称。
func ExeName() string {
	return strings.TrimSuffix(filepath.Base(exe), filepath.Ext(exe))
}

// ExeExt 返回包含前导点的可执行文件扩展名；没有扩展名时返回空字符串。
func ExeExt() string {
	return filepath.Ext(exe)
}

// WithExeExt 返回将可执行文件扩展名替换为 ext 后的完整路径。
// ext 缺少前导点时自动补齐；ext 为空时返回不含扩展名的路径。
func WithExeExt(ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return strings.TrimSuffix(exe, filepath.Ext(exe)) + ext
}
