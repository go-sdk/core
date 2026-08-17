package restx

import (
	"github.com/go-sdk/core/logx"
)

type logger struct{}

func (l *logger) Errorf(f string, v ...any) { logx.Error().Msgf(f, v...) }
func (l *logger) Warnf(f string, v ...any)  { logx.Warn().Msgf(f, v...) }
func (l *logger) Debugf(f string, v ...any) { logx.Debug().Msgf(f, v...) }
