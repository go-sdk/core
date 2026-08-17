package errx

import (
	"github.com/rotisserie/eris"
)

// New、Newf、Wrap、Wrapf、Unwrap、Cause、Is 和 As 是 eris 对应函数的别名。
var (
	New  = eris.New
	Newf = eris.Errorf

	Wrap  = eris.Wrap
	Wrapf = eris.Wrapf

	Unwrap = eris.Unwrap
	Cause  = eris.Cause

	Is = eris.Is
	As = eris.As
)

// Nil 是一个哨兵错误，用于在没有实际错误但需要非空错误值时占位。
var Nil = New("Nil")
