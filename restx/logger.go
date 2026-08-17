package restx

import (
	"github.com/go-sdk/core/logx"
)

// logger 将 resty 内部日志转发到 logx 全局日志。
type logger struct{}

func (l *logger) Errorf(f string, v ...any) { logx.Error().Msgf(f, v...) }
func (l *logger) Warnf(f string, v ...any)  { logx.Warn().Msgf(f, v...) }
func (l *logger) Debugf(f string, v ...any) { logx.Debug().Msgf(f, v...) }
