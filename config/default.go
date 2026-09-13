package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/spf13/cast"

	jsoncodec "github.com/go-sdk/core/codec/json"
	"github.com/go-sdk/core/osx"
)

var defaultConfig atomic.Pointer[Config]

func init() {
	filename, found, err := findDefaultFile()
	if err != nil {
		osx.Panicf("config: find default file: %v", err)
	}

	var c *Config
	if found {
		c = New(WithFile(filename))
	} else {
		c = New()
	}
	if err = c.Load(); err != nil {
		osx.Panicf("config: initialize default config: %v", err)
	}
	SetDefault(c)
}

// SetDefault 替换包级 Get 和 MustGet 使用的默认配置实例。
func SetDefault(c *Config) {
	if c == nil {
		osx.Panic("config: default config must not be nil")
	}
	defaultConfig.Store(c)
}

// Get 返回默认配置实例中的基础类型值。
func Get[T cast.Basic](key string) (T, bool) {
	return defaultConfig.Load().Get[T](key)
}

// MustGet 返回默认配置实例中的基础类型值，路径不存在时触发 panic。
func MustGet[T cast.Basic](key string) T {
	return defaultConfig.Load().MustGet[T](key)
}

func findDefaultFile() (string, bool, error) {
	// CONFIG_PATH 一经设置即直接采用且不回退，空值同样视为已设置。
	if filename, ok := os.LookupEnv("CONFIG_PATH"); ok {
		return filename, true, nil
	}
	if isTestProcess() {
		filename, err := testModuleConfigFile()
		if err != nil {
			return "", false, err
		}
		if filename != "" {
			return filename, true, nil
		}
	}
	for _, filename := range []string{
		osx.WithExeExt("yaml"),
		osx.WithExeExt("yml"),
		osx.WithExeExt("json"),
	} {
		exists, err := isRegularFile(filename)
		if err != nil {
			return "", false, err
		}
		if exists {
			return filename, true, nil
		}
	}
	return "", false, nil
}

func isTestProcess() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

func testModuleConfigFile() (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("go", "list", "-m", "-json")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run go list -m -json: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if message := strings.TrimSpace(stderr.String()); message != "" {
		return "", fmt.Errorf("go list -m -json wrote to stderr: %s", message)
	}
	var module struct {
		Dir string
	}
	if err := jsoncodec.Unmarshal(stdout.Bytes(), &module); err != nil {
		return "", fmt.Errorf("decode go list -m -json output: %w", err)
	}
	if module.Dir == "" {
		return "", fmt.Errorf("go list -m -json returned an empty module directory")
	}
	filename := filepath.Join(module.Dir, "config.yaml")
	exists, err := isRegularFile(filename)
	if err != nil || !exists {
		return "", err
	}
	return filename, nil
}

func isRegularFile(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat config file %q: %w", filename, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("config path %q is not a regular file", filename)
	}
	return true, nil
}
