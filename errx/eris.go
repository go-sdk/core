package errx

import (
	"errors"

	"github.com/rotisserie/eris"
)

// New、Newf、Wrap、Wrapf、Unwrap 和 Cause 是 eris 对应函数的别名。
// Join、Is 和 As 是标准库 errors 对应函数的别名，Is 和 As 可匹配 Join 产生的多错误链。
var (
	New  = eris.New
	Newf = eris.Errorf

	Wrap  = eris.Wrap
	Wrapf = eris.Wrapf

	Unwrap = eris.Unwrap
	Cause  = eris.Cause

	Join = errors.Join

	Is = errors.Is
	As = errors.As
)

// AsType 是标准库 errors.AsType 的入口，通过类型参数提取错误链中指定类型的错误。
func AsType[E error](err error) (E, bool) {
	return errors.AsType[E](err)
}

// Nil 是一个哨兵错误，用于在没有实际错误但需要非空错误值时占位。
var Nil = New("Nil")
