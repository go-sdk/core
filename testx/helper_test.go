package testx

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type aliasError struct{}

func (*aliasError) Error() string { return "别名错误" }

// mockTestingT 记录断言结果，避免失败分支终止当前测试。
type mockTestingT struct {
	failed      bool
	failNow     bool
	helperCalls int
	errors      []string
	logs        []string
}

func (m *mockTestingT) Errorf(format string, args ...any) {
	m.failed = true
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) FailNow() { m.failNow = true }
func (m *mockTestingT) Helper()  { m.helperCalls++ }
func (m *mockTestingT) Logf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}

func TestP(t *testing.T) {
	t.Run("没有参数时不记录内容", func(t *testing.T) {
		mock := &mockTestingT{}
		P(mock)

		if mock.failed || mock.failNow {
			t.Fatal("没有参数时不应失败")
		}
		if len(mock.logs) != 0 {
			t.Fatalf("没有参数时不应记录日志：%v", mock.logs)
		}
		if mock.helperCalls != 1 {
			t.Fatalf("应标记一次辅助函数，实际为 %d", mock.helperCalls)
		}
	})

	t.Run("忽略空错误并记录后续参数", func(t *testing.T) {
		mock := &mockTestingT{}
		var typedNil *aliasError

		P(mock, nil, "普通空值")
		P(mock, typedNil, "带类型空值")

		if mock.failed || mock.failNow {
			t.Fatal("空错误不应导致失败")
		}
		if len(mock.logs) != 2 {
			t.Fatalf("期望两条日志，实际为 %v", mock.logs)
		}
		if !strings.Contains(mock.logs[0], "普通空值") || !strings.Contains(mock.logs[1], "带类型空值") {
			t.Fatalf("日志内容不符合预期：%v", mock.logs)
		}
	})

	t.Run("非空错误立即失败", func(t *testing.T) {
		mock := &mockTestingT{}
		P(mock, errors.New("操作失败"))

		if !mock.failed || !mock.failNow {
			t.Fatalf("非空错误应触发失败：%+v", mock)
		}
		if len(mock.logs) != 0 {
			t.Fatalf("错误参数不应作为日志输出：%v", mock.logs)
		}
	})

	t.Run("格式化并逐个记录普通参数", func(t *testing.T) {
		mock := &mockTestingT{}
		P(mock, struct{ Value int }{Value: 7}, []string{"a", "b"})

		if mock.failed || mock.failNow {
			t.Fatal("普通参数不应导致失败")
		}
		if len(mock.logs) != 2 {
			t.Fatalf("期望两条日志，实际为 %v", mock.logs)
		}
		if !strings.Contains(mock.logs[0], "7") || !strings.Contains(mock.logs[1], "a") {
			t.Fatalf("格式化日志不符合预期：%v", mock.logs)
		}
	})
}

func TestAssertionAliases(t *testing.T) {
	NoError(t, nil)
	Error(t, errors.New("错误"))

	cause := errors.New("根因")
	wrapped := fmt.Errorf("外层：%w", cause)
	ErrorIs(t, wrapped, cause)
	EqualError(t, cause, "根因")
	ErrorContains(t, wrapped, "外层")

	typed := &aliasError{}
	var target *aliasError
	ErrorAs(t, fmt.Errorf("包装：%w", typed), &target)
	Same(t, typed, target)

	Equal(t, 1, 1)
	EqualValues(t, int32(1), int64(1))
	NotEqual(t, 1, 2)
	first := &struct{ Value int }{Value: 1}
	second := &struct{ Value int }{Value: 1}
	Same(t, first, first)
	NotSame(t, first, second)
	IsType(t, 0, 1)

	Nil(t, nil)
	NotNil(t, first)
	Empty(t, "")
	NotEmpty(t, "值")
	Zero(t, 0)
	NotZero(t, 1)
	True(t, true)
	False(t, false)

	Contains(t, "abcdef", "bcd")
	NotContains(t, "abcdef", "xyz")
	Len(t, []int{1, 2}, 2)
	ElementsMatch(t, []int{1, 2}, []int{2, 1})
	Greater(t, 2, 1)
	GreaterOrEqual(t, 2, 2)
	Less(t, 1, 2)
	LessOrEqual(t, 2, 2)
	Panics(t, func() { panic("预期异常") })
	NotPanics(t, func() {})

	mock := &mockTestingT{}
	Fail(mock, "预期失败")
	if !mock.failed || !mock.failNow {
		t.Fatalf("Fail 行为不符合预期：%+v", mock)
	}

	mock = &mockTestingT{}
	FailNow(mock, "预期立即失败")
	if !mock.failed || !mock.failNow {
		t.Fatalf("FailNow 行为不符合预期：%+v", mock)
	}
}
