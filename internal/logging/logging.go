package logging

import (
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	zerologlog "github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/go-sdk/core/osx"
)

const TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"

const consoleEnv = "APP__LOG__CONSOLE"

var configureOnce sync.Once

// Bootstrap 在配置加载前安装控制台日志，使启动阶段与运行阶段使用相同格式。
func Bootstrap() {
	logger := NewConsoleLogger(ConsoleOutput(), !ConsoleSupportsColor())
	Install(&logger)
}

// Configure 设置 zerolog 的进程级公共行为。
func Configure() {
	configureOnce.Do(func() {
		zerolog.TimeFieldFormat = TimeFieldFormat
		zerolog.ErrorStackMarshaler = func(err error) any {
			if err == nil {
				return nil
			}
			stack := strings.TrimSpace(eris.ToString(err, true))
			if osx.IsDebug() {
				_, _ = fmt.Fprintf(os.Stderr, "\n---\n%s [ERROR STACK]\n%s\n---\n", time.Now().Format(TimeFieldFormat), stack)
			}
			return stack
		}
		zerolog.ErrorMarshalFunc = zerolog.ErrorStackMarshaler
		zerolog.ErrorHandler = func(err error) { _, _ = fmt.Fprintf(os.Stderr, "non-expected logger error: %v", err) }
	})
}

// NewConsoleLogger 创建采用统一控制台格式的 zerolog Logger。
func NewConsoleLogger(out io.Writer, noColor bool) zerolog.Logger {
	return NewLogger(NewConsoleWriter(out, noColor))
}

// NewLogger 创建采用统一级别、调用位置和时间字段配置的 zerolog Logger。
func NewLogger(out io.Writer) zerolog.Logger {
	Configure()
	logger := zerolog.New(out).Level(zerolog.InfoLevel)
	if osx.IsDebug() {
		logger = logger.Level(zerolog.TraceLevel).With().Caller().Logger()
	}
	return logger.With().Timestamp().Logger()
}

// NewConsoleWriter 创建采用统一时间格式的控制台 Writer。
func NewConsoleWriter(out io.Writer, noColor bool) zerolog.ConsoleWriter {
	return zerolog.NewConsoleWriter(func(writer *zerolog.ConsoleWriter) {
		writer.Out = out
		writer.NoColor = noColor
		writer.TimeFormat = TimeFieldFormat
		writer.TimeLocation = time.Local
	})
}

// ConsoleOutput 根据 APP__LOG__CONSOLE 选择控制台输出，只有 stderr（忽略大小写）写入标准错误。
func ConsoleOutput() io.Writer {
	if consoleFile() == os.Stderr {
		return colorable.NewColorableStderr()
	}
	return colorable.NewColorableStdout()
}

// Install 让 zerolog、slog 和标准库 log 共享同一个全局 Logger。
func Install(logger *zerolog.Logger) {
	zerologlog.Logger = *logger
	zerolog.DefaultContextLogger = logger

	slogger := slog.New(newSlogHandler(logger))
	slog.SetDefault(slogger)

	standardLogger := slog.NewLogLogger(slogger.Handler(), slog.LevelInfo)
	stdlog.SetOutput(standardLogger.Writer())
	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
}

func newSlogHandler(logger *zerolog.Logger) slog.Handler {
	return slogzerolog.Option{Level: slog.LevelInfo, Logger: logger}.NewZerologHandler()
}

// StdoutSupportsColor 判断标准输出是否支持终端颜色。
func StdoutSupportsColor() bool {
	return fileSupportsColor(os.Stdout)
}

// ConsoleSupportsColor 判断当前控制台输出是否支持终端颜色。
func ConsoleSupportsColor() bool {
	return fileSupportsColor(consoleFile())
}

func consoleFile() *os.File {
	if strings.ToLower(os.Getenv(consoleEnv)) == "stderr" {
		return os.Stderr
	}
	return os.Stdout
}

func fileSupportsColor(file *os.File) bool {
	fd := file.Fd()
	return TerminalSupportsColor(os.Getenv("TERM"), isatty.IsTerminal(fd), isatty.IsCygwinTerminal(fd))
}

// TerminalSupportsColor 根据终端类型和能力判断是否支持颜色。
func TerminalSupportsColor(term string, terminal, cygwin bool) bool {
	return term != "dumb" && (terminal || cygwin)
}
