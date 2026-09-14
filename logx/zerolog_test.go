package logx

import (
	"bytes"
	"context"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	zerologlog "github.com/rs/zerolog/log"

	"github.com/go-sdk/core/testx"
)

type fieldHook struct{}

func (fieldHook) Run(event *zerolog.Event, _ zerolog.Level, _ string) {
	event.Str("hook", "已执行")
}

// useBufferLogger 将全局日志临时切换到内存缓冲区。
func useBufferLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	old := Logger
	buffer := &bytes.Buffer{}
	Logger = zerolog.New(buffer).Level(zerolog.TraceLevel)
	t.Cleanup(func() { Logger = old })
	return buffer
}

func TestEventFunctions(t *testing.T) {
	buffer := useBufferLogger(t)
	tests := []struct {
		name  string
		level string
		write func()
	}{
		{name: "Trace", level: "trace", write: func() { Trace().Msg("trace-message") }},
		{name: "Debug", level: "debug", write: func() { Debug().Msg("debug-message") }},
		{name: "Info", level: "info", write: func() { Info().Msg("info-message") }},
		{name: "Warn", level: "warn", write: func() { Warn().Msg("warn-message") }},
		{name: "Error", level: "error", write: func() { Error().Msg("error-message") }},
		{name: "ErrNil", level: "info", write: func() { Err(nil).Msg("nil-error-message") }},
		{name: "WithLevel", level: "warn", write: func() { WithLevel(zerolog.WarnLevel).Msg("level-message") }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer.Reset()
			test.write()
			output := buffer.String()
			testx.Contains(t, output, `"level":"`+test.level+`"`)
			testx.Contains(t, output, `"message":`)
		})
	}

	buffer.Reset()
	Log().Msg("no-level-message")
	testx.Contains(t, buffer.String(), `"message":"no-level-message"`)
	testx.NotContains(t, buffer.String(), `"level":`)
}

func TestDerivedLogger(t *testing.T) {
	globalBuffer := useBufferLogger(t)

	outputBuffer := &bytes.Buffer{}
	outputLogger := Output(outputBuffer)
	outputLogger.Info().Msg("独立输出")
	testx.Empty(t, globalBuffer.String())
	testx.Contains(t, outputBuffer.String(), "独立输出")

	globalBuffer.Reset()
	contextLogger := With().Str("component", "core").Logger()
	contextLogger.Info().Msg("带字段")
	testx.Contains(t, globalBuffer.String(), `"component":"core"`)

	globalBuffer.Reset()
	levelLogger := Level(zerolog.WarnLevel)
	levelLogger.Info().Msg("被过滤")
	levelLogger.Warn().Msg("被保留")
	testx.NotContains(t, globalBuffer.String(), "被过滤")
	testx.Contains(t, globalBuffer.String(), "被保留")

	globalBuffer.Reset()
	hookLogger := Hook(fieldHook{})
	hookLogger.Info().Msg("带钩子")
	testx.Contains(t, globalBuffer.String(), `"hook":"已执行"`)

	globalBuffer.Reset()
	sampleLogger := Sample(nil)
	sampleLogger.Info().Msg("不采样")
	testx.Contains(t, globalBuffer.String(), "不采样")
}

func TestPrintFunctions(t *testing.T) {
	buffer := useBufferLogger(t)

	Print("值", 1)
	Printf("格式化-%d", 2)

	output := buffer.String()
	testx.Equal(t, 2, strings.Count(output, "\n"))
	testx.Contains(t, output, `"message":"值1"`)
	testx.Contains(t, output, `"message":"格式化-2"`)
}

func TestContextLogger(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := zerolog.New(buffer)
	ctx := logger.WithContext(context.Background())

	Ctx(ctx).Info().Msg("上下文日志")

	testx.Contains(t, buffer.String(), "上下文日志")
	testx.NotNil(t, Ctx(context.Background()))
}

func TestNewLogger(t *testing.T) {
	logger := New("")

	testx.Equal(t, zerolog.TraceLevel, logger.GetLevel())
	outputLogger := logger.Output(io.Discard)
	outputLogger.Info().Msg("可写")
}

func TestTerminalSupportsColor(t *testing.T) {
	testx.True(t, terminalSupportsColor("xterm-256color", true, false))
	testx.True(t, terminalSupportsColor("xterm-256color", false, true))
	testx.False(t, terminalSupportsColor("xterm-256color", false, false))
	testx.False(t, terminalSupportsColor("dumb", true, true))
}

func TestInitReconfiguresGlobalLogger(t *testing.T) {
	firstFile := filepath.Join(t.TempDir(), "first.log")
	secondFile := filepath.Join(t.TempDir(), "second.log")
	t.Cleanup(func() { Init("") })

	Init(firstFile)
	Info().Msg("第一阶段")
	zerologlog.Info().Msg("zerolog 全局日志")
	slog.Info("slog 默认日志")
	stdlog.Print("标准库日志")

	Init(secondFile)
	Info().Msg("第二阶段")
	Close()

	first, err := os.ReadFile(firstFile)
	testx.NoError(t, err)
	firstOutput := string(first)
	testx.Contains(t, firstOutput, "第一阶段")
	testx.Contains(t, firstOutput, "zerolog 全局日志")
	testx.Contains(t, firstOutput, "slog 默认日志")
	testx.Contains(t, firstOutput, "标准库日志")
	testx.NotContains(t, firstOutput, "第二阶段")

	second, err := os.ReadFile(secondFile)
	testx.NoError(t, err)
	testx.Contains(t, string(second), "第二阶段")
}
