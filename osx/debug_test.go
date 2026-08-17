package osx

import (
	"os"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestIsDebug(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	t.Run("测试参数启用调试模式", func(t *testing.T) {
		os.Args = []string{"app", "-test.v"}
		t.Setenv("DEBUG", "false")
		testx.True(t, IsDebug())
	})

	t.Run("普通参数不启用调试模式", func(t *testing.T) {
		os.Args = []string{"app", "--test.mode"}
		t.Setenv("DEBUG", "false")
		testx.False(t, IsDebug())
	})

	truthyValues := []string{"1", "t", "true", "y", "yes", "on", "TRUE", "On"}
	for _, value := range truthyValues {
		t.Run("环境变量_"+value, func(t *testing.T) {
			os.Args = []string{"app"}
			t.Setenv("DEBUG", value)
			testx.True(t, IsDebug())
		})
	}

	t.Run("其他环境变量值不启用调试模式", func(t *testing.T) {
		os.Args = []string{"app"}
		t.Setenv("DEBUG", "enabled")
		testx.False(t, IsDebug())
	})
}
