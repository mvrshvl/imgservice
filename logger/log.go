package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
)

func NewLogFile(logPath string) (io.Writer, error) {
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return os.Create(filepath.Join(logPath, fmt.Sprintf("%s.log", time.Now().String())))
}

func New(logLevel logrus.Level, output io.Writer) *logrus.Entry {
	logger := logrus.New()

	logger.SetOutput(output)
	logger.SetLevel(logLevel)

	return logrus.NewEntry(logger)
}

func WithCtx(ctx context.Context, log *logrus.Entry) context.Context {
	return ctxlogrus.ToContext(ctx, log)
}

func Error(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Error(args...)
}

func Info(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Info(args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Debug(args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Infof(format, args...)
}

func Warn(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Warn(args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Warnf(format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.WithContext(ctx).Errorf(format, args...)
}
