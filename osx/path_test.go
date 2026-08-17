package osx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestWorkDir(t *testing.T) {
	want, err := os.Getwd()
	testx.NoError(t, err)
	testx.Equal(t, want, WorkDir())
}

func TestExecutablePath(t *testing.T) {
	full := ExeFull()
	testx.NotEmpty(t, full)
	testx.True(t, filepath.IsAbs(full))
	testx.Equal(t, filepath.Dir(full), ExeDir())
	testx.Equal(t, filepath.Ext(full), ExeExt())
	testx.Equal(t, strings.TrimSuffix(filepath.Base(full), filepath.Ext(full)), ExeName())
}

func TestWithExeExt(t *testing.T) {
	base := strings.TrimSuffix(ExeFull(), filepath.Ext(ExeFull()))

	t.Run("缺少前导点时自动补齐", func(t *testing.T) {
		testx.Equal(t, base+".log", WithExeExt("log"))
	})

	t.Run("保留已有前导点", func(t *testing.T) {
		testx.Equal(t, base+".log", WithExeExt(".log"))
	})

	t.Run("空扩展名返回无扩展名路径", func(t *testing.T) {
		testx.Equal(t, base, WithExeExt(""))
	})
}
