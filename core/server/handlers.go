package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"imgservice/core/fs"
	"imgservice/html"
)

func HomeGetHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.Home))
}

func ResizeGetHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.Resize))
}

func ResizePostHandler(ctx *gin.Context) {
	postHandler(ctx, resize)
}

func ResizePercentGetHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.ResizePercent))
}

func ResizePercentPostHandler(ctx *gin.Context) {
	postHandler(ctx, resizePercent)
}

func GrayScaleGetHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.GrayScale))
}

func GrayPostHandler(ctx *gin.Context) {
	postHandler(ctx, grayScaling)
}

func WatermarkGetHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.Watermark))
}

func WatermarkPostHandler(ctx *gin.Context) {
	postHandler(ctx, addWatermark)
}

func ConvertGetHandler(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.Convert))
}

func ConvertPostHandler(ctx *gin.Context) {
	postHandler(ctx, convertExt)
}

func DownloadGetHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	storage, err := fs.GetCtx(ctx)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	ctx.FileFromFS(id, storage)
}

func postHandler(ctx *gin.Context, prepareResponse func(ctx *gin.Context) (string, error)) {
	id, err := prepareResponse(ctx)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))

		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(html.Download, id)))
}
