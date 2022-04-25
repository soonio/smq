package smq

import (
	"go.uber.org/zap"
)

type zapLogAdapter struct {
	ZapLogger *zap.Logger
}

func (l *zapLogAdapter) Info(msg string) {
	l.logger().Sugar().Info(msg)
}
func (l *zapLogAdapter) Error(msg string) {
	l.logger().Sugar().Error(msg)
}

func Zap2Log(zap *zap.Logger) *zapLogAdapter {
	return &zapLogAdapter{zap}
}

func (l *zapLogAdapter) logger() *zap.Logger {
	return l.ZapLogger.WithOptions(zap.AddCallerSkip(1))
}
