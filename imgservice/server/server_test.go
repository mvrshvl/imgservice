package server

import (
	"bytes"
	"context"
	"errors"
	"imgservice/logger"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestServer_StartStop(t *testing.T) {
	router := gin.New()

	server := New(":6565", router, time.Second)
	errCh := make(chan error, 1)

	buff := new(bytes.Buffer)

	log := logger.New(logrus.DebugLevel, buff)

	ctx := context.Background()
	ctx = logger.WithCtx(ctx, log)

	go server.Run(errCh)

	server.Stop(ctx)

	resultLog, err := io.ReadAll(buff)
	if err != nil {
		t.Fatal(err)
	}

	if len(resultLog) > 0 {
		t.Fatalf("unexpected error: %v", string(resultLog))
	}

	err = <-errCh
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("expected %v, result %v", err, http.ErrServerClosed)
	}
}
