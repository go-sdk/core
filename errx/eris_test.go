package errx

import (
	"errors"
	"testing"

	"github.com/go-sdk/core/testx"
)

type testError struct {
	code int
}

func (e *testError) Error() string { return "自定义错误" }

func TestCreateError(t *testing.T) {
	t.Run("固定消息", func(t *testing.T) {
		err := New("操作失败")
		testx.EqualError(t, err, "操作失败")
	})

	t.Run("格式化消息", func(t *testing.T) {
		err := Newf("状态码：%d", 42)
		testx.EqualError(t, err, "状态码：42")
	})
}

func TestWrapError(t *testing.T) {
	cause := errors.New("根因")
	err := Wrap(cause, "第一层")
	err = Wrapf(err, "第%d层", 2)

	testx.ErrorIs(t, err, cause)
	testx.Equal(t, cause, Cause(err))
	testx.NotEqual(t, cause, Unwrap(err))
	testx.Contains(t, err.Error(), "第2层")
	testx.Contains(t, err.Error(), "第一层")

	t.Run("空错误保持为空", func(t *testing.T) {
		testx.NoError(t, Wrap(nil, "上下文"))
		testx.NoError(t, Wrapf(nil, "上下文：%d", 1))
		testx.NoError(t, Cause(nil))
	})
}

func TestMatchErrorType(t *testing.T) {
	source := &testError{code: 7}
	err := Wrap(source, "外层")

	var target *testError
	testx.True(t, As(err, &target))
	testx.Same(t, source, target)
	testx.Equal(t, 7, target.code)

	var other interface{ Temporary() bool }
	testx.False(t, As(err, &other))
	testx.False(t, Is(err, errors.New("其他错误")))

	t.Run("AsType 按类型提取", func(t *testing.T) {
		matched, ok := AsType[*testError](err)
		testx.True(t, ok)
		testx.Same(t, source, matched)

		_, ok = AsType[*testError](errors.New("其他错误"))
		testx.False(t, ok)

		_, ok = AsType[*testError](nil)
		testx.False(t, ok)
	})
}

func TestJoinError(t *testing.T) {
	first := errors.New("第一个错误")
	second := errors.New("第二个错误")

	t.Run("全部为空返回空", func(t *testing.T) {
		testx.NoError(t, Join())
		testx.NoError(t, Join(nil, nil))
	})

	t.Run("保留全部错误消息", func(t *testing.T) {
		testx.EqualError(t, Join(first, nil, second), "第一个错误\n第二个错误")
	})

	t.Run("匹配任意成员", func(t *testing.T) {
		err := Join(first, second)
		testx.ErrorIs(t, err, first)
		testx.ErrorIs(t, err, second)
		testx.False(t, Is(err, errors.New("其他错误")))
	})

	t.Run("包装过的成员同样可以匹配", func(t *testing.T) {
		cause := errors.New("根因")
		err := Join(Wrap(cause, "外层"), second)
		testx.ErrorIs(t, err, cause)
	})

	t.Run("按类型提取成员", func(t *testing.T) {
		source := &testError{code: 9}
		err := Join(first, Wrap(source, "外层"))

		var target *testError
		testx.True(t, As(err, &target))
		testx.Same(t, source, target)
	})
}

func TestNilSentinel(t *testing.T) {
	testx.NotNil(t, Nil)
	testx.EqualError(t, Nil, "Nil")
	testx.True(t, Is(Nil, Nil))
}
