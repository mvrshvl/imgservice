package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"imgservice/config"
	"imgservice/core/fs"
	"imgservice/core/server"
	"imgservice/logger"
)

type Runner struct {
	server *server.Server

	err chan error
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, cfg *config.Config) (context.Context, error) {
	serverCfg := cfg.GetServer()

	ctx, err := setLogger(ctx, serverCfg.LogLevel, serverCfg.LogPath)
	if err != nil {
		return nil, err
	}

	r.runServer(ctx, serverCfg)

	return ctx, nil
}

func (r *Runner) runServer(ctx context.Context, cfg config.Server) {
	router := gin.New()

	setMiddlewares(router, fs.New())

	server.SetGET(router)
	server.SetPOST(router)

	r.server = server.New(cfg.Address, router, cfg.Timeout)
	r.err = make(chan error, 1)

	logger.Info(ctx, "Run server...")

	go r.server.Run(r.err)
}

func (r *Runner) Wait(ctx context.Context) {
	defer r.gracefulShutdown(ctx)

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGHUP,
		syscall.SIGTERM,
	)

	select {
	case sig := <-signalChannel:
		logger.Warnf(ctx, "signal: %s, server shutdown.", sig)

		return
	case err := <-r.err:
		logger.Errorf(ctx, "server stopped with error %v", err)

		return
	}
}

func (r *Runner) gracefulShutdown(ctx context.Context) {
	r.server.Stop(ctx)
}

func setLogger(ctx context.Context, logLevel logrus.Level, logPath string) (context.Context, error) {
	var logWriter io.Writer = os.Stdout

	if len(logPath) != 0 {
		logFile, err := logger.NewLogFile(logPath)
		if err != nil {
			return nil, fmt.Errorf("can't create log file: %w", err)
		}

		logWriter = logFile
	}

	logEntry := logger.New(logLevel, logWriter)
	ctxWithLogger := logger.WithCtx(ctx, logEntry)

	return ctxWithLogger, nil
}

func setMiddlewares(router *gin.Engine, strg fs.InMemoryFS) {
	router.Use(gin.Recovery(), fs.SetCtx(strg))
}
