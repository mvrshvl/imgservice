package server

import (
	"context"
	"imgservice/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	http.Server
	sync.Once
}

func New(address string, handler http.Handler, timeout time.Duration) *Server {
	return &Server{
		Server: http.Server{
			Addr:         address,
			Handler:      handler,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (server *Server) Run(errChan chan error) {
	err := server.ListenAndServe()
	if err != nil {
		errChan <- err
	}
}

func (server *Server) Stop(ctx context.Context) {
	server.Do(func() {
		err := server.Shutdown(ctx)
		if err != nil {
			logger.Error(ctx, err)
		}
	})
}

func SetGET(router *gin.Engine) {
	router.GET(home, Home)
	router.GET(resize, Resize)
	router.GET(resizePercent, ResizePercent)
	router.GET(grayscale, Gray)
	router.GET(convert, Convert)
	router.GET(watermark, Watermark)
	router.GET(download, Download)
}

func SetPOST(router *gin.Engine) {
	router.POST(resizePercentLoad, ResizePercentPOST)
}
