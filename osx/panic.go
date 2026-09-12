package osx

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Panic 先将给定值和从调用方开始的调用堆栈以易读格式打印到标准错误输出，再调用 panic 保持原有行为。
func Panic(v any) {
	panicWithStack(3, v)
}

// Panicf 按 format 格式化参数后，行为与 Panic 一致。
func Panicf(format string, args ...any) {
	panicWithStack(3, fmt.Sprintf(format, args...))
}

// panicWithStack 打印从第 skip 层调用方开始的堆栈后调用 panic；skip 与 runtime.Callers 的参数语义一致。
func panicWithStack(skip int, v any) {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(skip, pcs)

	var b strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	for i := 1; ; i++ {
		frame, more := frames.Next()
		name := frame.Function
		if name == "" {
			name = "unknown"
		}
		fmt.Fprintf(&b, "%2d. %s\n    %s:%d", i, name, frame.File, frame.Line)
		if !more {
			break
		}
		b.WriteByte('\n')
	}

	_, _ = fmt.Fprintf(os.Stderr, "\n---\n%s [PANIC]\n%v\n%s\n---\n", time.Now().Format("2006-01-02T15:04:05.000Z07:00"), v, b.String())
	panic(v)
}
