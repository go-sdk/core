package cmdx

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/go-sdk/core/errx"
	"github.com/go-sdk/core/osx"
)

// Command 是 cobra.Command 的类型别名。
type Command = cobra.Command

var (
	// NoArgs 校验命令不接受位置参数。
	NoArgs = cobra.NoArgs
	// OnlyValidArgs 校验所有位置参数都出现在命令的 ValidArgs 中。
	OnlyValidArgs = cobra.OnlyValidArgs
	// ArbitraryArgs 允许任意数量的位置参数，始终返回 nil。
	ArbitraryArgs = cobra.ArbitraryArgs
	// MinimumNArgs 校验位置参数数量不少于 n 个。
	MinimumNArgs = cobra.MinimumNArgs
	// MaximumNArgs 校验位置参数数量不超过 n 个。
	MaximumNArgs = cobra.MaximumNArgs
	// ExactArgs 校验位置参数数量恰好为 n 个。
	ExactArgs = cobra.ExactArgs
	// RangeArgs 校验位置参数数量位于 min 和 max 的闭区间内。
	RangeArgs = cobra.RangeArgs
	// MatchAll 按顺序组合多个位置参数校验器，任一校验失败即返回错误。
	MatchAll = cobra.MatchAll
)

// NewRoot 创建带统一默认配置的根命令。
// 隐藏 help 和 completion 子命令，版本描述来自 osx.GetVersion 并以单行输出；
// 帮助与版本写入标准输出，cobra 自身的错误输出被丢弃，
// 命令错误仍由 Execute 返回，由调用方决定如何记录。
func NewRoot(name string) *Command {
	c := &cobra.Command{Use: name}
	c.SetHelpCommand(&cobra.Command{Hidden: true})
	c.CompletionOptions.HiddenDefaultCmd = true
	c.SetVersionTemplate("{{ print .Version }}")
	x := osx.GetVersion()
	c.Version = fmt.Sprintf("%s %s (%s) at %s built %s\n", c.Use, x.Version, x.Revision, x.Platform, x.Time)
	c.SetOut(os.Stdout)
	c.SetErr(io.Discard)
	return c
}

// WrapRunE 包装 RunE：errx.Nil 哨兵错误视为无错误并返回 nil，
// 其余错误原样返回，避免占位错误被当作真实错误触发 cobra 的错误输出和用法提示。
func WrapRunE(fn func(*Command, []string) error) func(*Command, []string) error {
	return func(cmd *Command, args []string) error {
		if err := fn(cmd, args); err != nil && !errx.Is(err, errx.Nil) {
			return err
		}
		return nil
	}
}
