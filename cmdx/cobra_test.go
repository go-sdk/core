package cmdx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/core/testx"
)

func TestNewRoot(t *testing.T) {
	root := NewRoot("app")

	testx.Equal(t, "app", root.Use)
	testx.True(t, root.CompletionOptions.HiddenDefaultCmd)
	testx.Equal(t, "{{ print .Version }}", root.VersionTemplate())

	x := osx.GetVersion()
	testx.Equal(t,
		fmt.Sprintf("%s %s (%s) at %s built %s\n", root.Use, x.Version, x.Revision, x.Platform, x.Time),
		root.Version)
	testx.Equal(t, os.Stdout, root.OutOrStdout())
	testx.Equal(t, io.Discard, root.ErrOrStderr())
}

func TestExecuteVersionAndHelp(t *testing.T) {
	var out bytes.Buffer

	root := NewRoot("app")
	root.AddCommand(&Command{Use: "sub", Run: func(*Command, []string) {}})
	root.SetOut(&out)
	root.SetArgs([]string{"--version"})

	testx.NoError(t, root.Execute())
	testx.Equal(t, root.Version, out.String())

	out.Reset()
	root.SetArgs([]string{"--help"})

	testx.NoError(t, root.Execute())
	testx.Contains(t, out.String(), "Available Commands:")
	testx.Contains(t, out.String(), "sub")
	testx.NotContains(t, out.String(), "Help about any command")
	testx.NotContains(t, out.String(), "completion")
}

func TestWrapRunE(t *testing.T) {
	root := NewRoot("app")

	t.Run("无错误时返回 nil 并透传参数", func(t *testing.T) {
		var gotCmd *Command
		var gotArgs []string
		wrapped := WrapRunE(func(cmd *Command, args []string) error {
			gotCmd, gotArgs = cmd, args
			return nil
		})

		testx.NoError(t, wrapped(root, []string{"a", "b"}))
		testx.Same(t, root, gotCmd)
		testx.Equal(t, []string{"a", "b"}, gotArgs)
	})

	t.Run("errx.Nil 视为无错误", func(t *testing.T) {
		wrapped := WrapRunE(func(cmd *Command, args []string) error { return errx.Nil })
		testx.NoError(t, wrapped(root, nil))
	})

	t.Run("包装过的 errx.Nil 同样视为无错误", func(t *testing.T) {
		wrapped := WrapRunE(func(cmd *Command, args []string) error {
			return errx.Wrap(errx.Nil, "占位")
		})
		testx.NoError(t, wrapped(root, nil))
	})

	t.Run("真实错误原样返回", func(t *testing.T) {
		target := errx.New("执行失败")
		wrapped := WrapRunE(func(cmd *Command, args []string) error { return target })
		testx.Same(t, target, wrapped(root, nil))
	})
}
