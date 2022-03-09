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
	router.GET(homePath, HomeGetHandler)
	router.GET(resizePath, ResizeGetHandler)
	router.GET(resizePercentPath, ResizePercentGetHandler)
	router.GET(grayscalePath, GrayScaleGetHandler)
	router.GET(convertPath, ConvertGetHandler)
	router.GET(watermarkPath, WatermarkGetHandler)
	router.GET(downloadPath, DownloadGetHandler)
}

func SetPOST(router *gin.Engine) {
	router.POST(resizePercentLoadPath, ResizePercentPostHandler)
	router.POST(resizePathLoad, ResizePostHandler)
	router.POST(grayscalePathLoad, GrayPostHandler)
	router.POST(convertPathLoad, ConvertPostHandler)
	router.POST(watermarkPathLoad, WatermarkPostHandler)
}
