package testx

import (
	"reflect"
	"strings"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
)

// Error、Equal、Nil、True、Contains、Len、Greater、Panics、Fail
// 及下列其他断言函数是 testify require 对应函数的别名。
var (
	NoError       = require.NoError
	Error         = require.Error
	ErrorIs       = require.ErrorIs
	ErrorAs       = require.ErrorAs
	EqualError    = require.EqualError
	ErrorContains = require.ErrorContains

	Equal       = require.Equal
	EqualValues = require.EqualValues
	NotEqual    = require.NotEqual
	Same        = require.Same
	NotSame     = require.NotSame
	IsType      = require.IsType

	Nil      = require.Nil
	NotNil   = require.NotNil
	Empty    = require.Empty
	NotEmpty = require.NotEmpty
	Zero     = require.Zero
	NotZero  = require.NotZero

	True  = require.True
	False = require.False

	Contains      = require.Contains
	NotContains   = require.NotContains
	Len           = require.Len
	ElementsMatch = require.ElementsMatch

	Greater        = require.Greater
	GreaterOrEqual = require.GreaterOrEqual
	Less           = require.Less
	LessOrEqual    = require.LessOrEqual

	Panics    = require.Panics
	NotPanics = require.NotPanics
	Fail      = require.Fail
	FailNow   = require.FailNow
)

// TestingT 定义 P 所需的测试上下文能力。
type TestingT interface {
	require.TestingT
	Helper()
	Logf(format string, args ...any)
}

// P 检查首个参数中的错误并使用易读格式记录其余参数。
// 首个参数为 nil 或带类型的 nil 时会被忽略；首个参数为非空错误时立即终止当前测试。
func P(t TestingT, args ...any) {
	t.Helper()

	if len(args) == 0 {
		return
	}

	if isNilValue(args[0]) {
		args = args[1:]
	} else if e, ok := args[0].(error); ok {
		NoError(t, e)
		args = args[1:]
	}

	if len(args) == 0 {
		return
	}

	for _, arg := range args {
		if s := pretty.Sprint(arg); strings.Contains(s, "\n") {
			t.Logf("\n%s", s)
		} else {
			t.Logf("%s", s)
		}
	}
}

// isNilValue 判断 v 是否为 nil，同时识别 nil *MyError 一类带类型的 nil 值。
func isNilValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
