package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"imgservice/archive"
	"imgservice/fs"
	"imgservice/html"
	"imgservice/image"
	"mime/multipart"
	"net/http"
	"strconv"
)

func Home(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.Home))
}

func Resize(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte("resize page"))
}

type ResizePercentRequest struct {
	Size string                `form:"size"`
	File *multipart.FileHeader `form:"file"`
}

func ResizePercent(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html.Resize))
}

func ResizePercentPOST(ctx *gin.Context) {
	var resizeRequest ResizePercentRequest

	err := ctx.ShouldBind(&resizeRequest)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	percent, err := strconv.ParseFloat(resizeRequest.Size, 64)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	f, err := resizeRequest.File.Open()
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	imgs, err := archive.Unzip(f, resizeRequest.File.Size)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	resizedIMGs := make([]*image.Image, len(imgs))
	for i, img := range imgs {
		resizedIMG, err := img.ResizePercent(percent)
		if err != nil {
			ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
			return
		}

		resizedIMGs[i] = resizedIMG
	}

	resizedArchive, err := archive.Zip(resizedIMGs)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	id, err := saveFile(ctx, resizedArchive)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(html.Download, id)))
}

func Gray(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte("gray scaling page"))
}

func Watermark(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte("watermark page"))
}

func Convert(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte("convert page"))
}

func Download(ctx *gin.Context) {
	id := ctx.Param("id")

	storage, err := fs.GetCtx(ctx)
	if err != nil {
		ctx.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(err.Error()))
		return
	}

	ctx.FileFromFS(id, storage)
}

func saveFile(ctx *gin.Context, raw []byte) (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	storage, err := fs.GetCtx(ctx)
	if err != nil {
		return "", err
	}

	storage.Add(id.String(), raw)

	return id.String(), nil
}
