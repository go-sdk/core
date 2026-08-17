package osx

import (
	"os"
	"testing"
	"time"

	"github.com/go-sdk/core/testx"
)

func TestGetEnv(t *testing.T) {
	const (
		first  = "GO_SDK_CORE_TEST_ENV_FIRST"
		second = "GO_SDK_CORE_TEST_ENV_SECOND"
	)

	t.Run("全部未设置时返回默认值", func(t *testing.T) {
		unsetEnv(t, first)
		unsetEnv(t, second)
		testx.Equal(t, 42, GetEnv(42, first, second))
	})

	t.Run("使用第一个已设置变量", func(t *testing.T) {
		t.Setenv(first, "11")
		t.Setenv(second, "22")
		testx.Equal(t, 11, GetEnv(99, first, second))
	})

	t.Run("使用后续已设置变量", func(t *testing.T) {
		unsetEnv(t, first)
		t.Setenv(second, "22")
		testx.Equal(t, 22, GetEnv(99, first, second))
	})

	t.Run("空值不回退", func(t *testing.T) {
		t.Setenv(first, "")
		t.Setenv(second, "second")
		testx.Equal(t, "", GetEnv("default", first, second))
	})

	t.Run("转换失败不回退", func(t *testing.T) {
		t.Setenv(first, "invalid")
		t.Setenv(second, "42")
		testx.Equal(t, 0, GetEnv(100, first, second))
	})

	t.Run("支持基础类型转换", func(t *testing.T) {
		t.Setenv(first, "true")
		testx.True(t, GetEnv(false, first))

		t.Setenv(first, "1.5s")
		testx.Equal(t, 1500*time.Millisecond, GetEnv(time.Duration(0), first))
	})
}

// unsetEnv 确保环境变量未设置，并在测试结束后恢复原始状态。
func unsetEnv(t *testing.T, name string) {
	t.Helper()
	value, ok := os.LookupEnv(name)
	testx.NoError(t, os.Unsetenv(name))
	t.Cleanup(func() {
		if ok {
			testx.NoError(t, os.Setenv(name, value))
		}
	})
}
