package logx

import (
	"io"
	"sync"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/go-sdk/core/config"
	"github.com/go-sdk/core/internal/logging"
	"github.com/go-sdk/core/lifex"
)

// ZeroLogger 是 zerolog.Logger 的类型别名。
type ZeroLogger = zerolog.Logger

func init() {
	Init(config.MustGet[string]("log.file.path", ""))
	lifex.OnDeinit(Close)
}

// Init 初始化或重新配置全局日志，并同步设置 zerolog 包级日志、slog 默认日志和标准库日志。
// 日志始终输出到标准输出；filename 非空时同时写入滚动文件。
// 程序通常先由包初始化建立默认日志，再在读取配置后调用一次 Init。
// 重新初始化前会关闭此前创建的文件 Writer；使用 lifex 时由进程生命周期自动关闭。
func Init(filename string) {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()
	closeFileWriters()
	install(New(filename))
}

// New 创建写入标准输出的 zerolog 日志；filename 非空时同时写入滚动文件。
// 调试模式下最低级别为 trace，并记录调用位置。
// 本包按进程维护统一的文件 Writer 生命周期，调用 Init 或 Close 会关闭此前创建的文件 Writer。
func New(filename string) zerolog.Logger {
	return logging.NewLogger(loggerWriters(filename)).Hook(globalKvHook{})
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
	return logging.NewConsoleWriter(
		colorable.NewColorableStdout(),
		config.MustGet[bool]("log.no_color", !logging.StdoutSupportsColor()),
	)
}

func loggerFileWriter(filename string) io.WriteCloser {
	if filename == "" {
		return nil
	}
	fw := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MustGet[int]("log.file.size", 30),
		MaxAge:     config.MustGet[int]("log.file.age", 90),
		MaxBackups: config.MustGet[int]("log.file.backups", 10),
		LocalTime:  config.MustGet[bool]("log.file.localtime", true),
		Compress:   config.MustGet[bool]("log.file.compress", true),
	}
	w := diode.NewWriter(fw, 1000, 10*time.Millisecond, nil)
	writersMu.Lock()
	wcs = append(wcs, w)
	writersMu.Unlock()
	return w
}

var (
	wcs         []io.WriteCloser
	writersMu   sync.Mutex
	lifecycleMu sync.Mutex
)

// Close 将全局日志切回控制台，并刷新、关闭由 New 创建的全部文件 Writer。
func Close() {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()
	install(New(""))
	closeFileWriters()
}

func install(logger zerolog.Logger) {
	Logger = logger
	logging.Install(&Logger)
}

func closeFileWriters() {
	writersMu.Lock()
	writers := wcs
	wcs = nil
	writersMu.Unlock()
	for _, wc := range writers {
		_ = wc.Close()
	}
}
