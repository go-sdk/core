package logx

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
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/go-sdk/core/osx"
)

// ZeroLogger 是 zerolog.Logger 的类型别名。
type ZeroLogger = zerolog.Logger

var (
	logs *slog.Logger
	loge *stdlog.Logger
)

func init() {
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	zerolog.ErrorStackMarshaler = func(e error) any {
		if e == nil {
			return nil
		}
		s := strings.TrimSpace(eris.ToString(e, true))
		if osx.IsDebug() {
			_, _ = fmt.Fprintf(os.Stderr, "\n---\n%s [ERROR STACK]\n%s\n---\n", time.Now().Format(zerolog.TimeFieldFormat), s)
		}
		return s
	}
	zerolog.ErrorMarshalFunc = zerolog.ErrorStackMarshaler
	zerolog.ErrorHandler = func(err error) { _, _ = fmt.Fprintf(os.Stderr, "non-expected logger error: %v", err) }

	Init(osx.GetEnv[string]("", "LOGX_FILE_PATH"))
}

// Init 初始化或重新配置全局日志，并同步设置 zerolog 包级日志、slog 默认日志和标准库日志。
// 日志始终输出到标准输出；filename 非空时同时写入滚动文件。
// 程序通常先由包初始化建立默认日志，再在读取配置后调用一次 Init。
// 重新初始化前会关闭此前创建的文件 Writer，程序退出前应调用 Close 完成刷新和关闭。
func Init(filename string) {
	Close()
	Logger = New(filename)
	log.Logger = Logger
	zerolog.DefaultContextLogger = &Logger

	logs = slog.New(slogzerolog.Option{Level: slog.LevelInfo, Logger: &Logger}.NewZerologHandler())
	slog.SetDefault(logs)

	loge = slog.NewLogLogger(logs.Handler(), slog.LevelInfo)
	stdlog.SetOutput(loge.Writer())
	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
}

// New 创建写入标准输出的 zerolog 日志；filename 非空时同时写入滚动文件。
// 调试模式下最低级别为 trace，并记录调用位置。
// 本包按进程维护统一的文件 Writer 生命周期，调用 Init 或 Close 会关闭此前创建的文件 Writer。
func New(filename string) zerolog.Logger {
	zl := zerolog.New(loggerWriters(filename)).Level(zerolog.InfoLevel).Hook(globalKvHook{})
	if osx.IsDebug() {
		zl = zl.Level(zerolog.TraceLevel).With().Caller().Logger()
	}
	return zl.With().Timestamp().Logger()
}

func loggerWriters(filename string) io.Writer {
	ws1 := []io.WriteCloser{
		loggerConsoleWriter(),
		loggerFileWriter(filename),
	}
	ws2 := make([]io.Writer, 0, len(ws1))
	for _, w := range ws1 {
		if w != nil && w != io.Discard {
			ws2 = append(ws2, w)
		}
	}
	return zerolog.MultiLevelWriter(ws2...)
}

func loggerConsoleWriter() io.WriteCloser {
	return zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = colorable.NewColorableStdout()
		w.NoColor = osx.GetEnv[bool](false, "LOGX_NO_COLOR")
		w.TimeFormat = zerolog.TimeFieldFormat
		w.TimeLocation = time.Local
	})
}

func loggerFileWriter(filename string) io.WriteCloser {
	if filename == "" {
		return nil
	}
	fw := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    osx.GetEnv[int](30, "LOGX_FILE_SIZE"),
		MaxAge:     osx.GetEnv[int](90, "LOGX_FILE_AGE"),
		MaxBackups: osx.GetEnv[int](10, "LOGX_FILE_BACKUPS"),
		LocalTime:  osx.GetEnv[bool](true, "LOGX_FILE_LOCALTIME"),
		Compress:   osx.GetEnv[bool](true, "LOGX_FILE_COMPRESS"),
	}
	w := diode.NewWriter(fw, 1000, 10*time.Millisecond, nil)
	mu.Lock()
	wcs = append(wcs, w)
	mu.Unlock()
	return w
}

var (
	wcs []io.WriteCloser
	mu  sync.Mutex
)

// Close 刷新并关闭由 New 创建的全部文件 Writer。
func Close() {
	mu.Lock()
	defer mu.Unlock()
	for _, wc := range wcs {
		_ = wc.Close()
	}
	wcs = wcs[:0]
}
