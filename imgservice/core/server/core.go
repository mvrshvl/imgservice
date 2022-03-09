package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"imgservice/core/archive"
	"imgservice/core/fs"
	"imgservice/core/image"
	"mime/multipart"
	"strconv"
)

func resize(ctx *gin.Context) (string, error) {
	var resizeRequest ResizeRequest

	err := ctx.ShouldBind(&resizeRequest)
	if err != nil {
		return "", err
	}

	height, err := strconv.Atoi(resizeRequest.Height)
	if err != nil {
		return "", err
	}

	width, err := strconv.Atoi(resizeRequest.Width)
	if err != nil {
		return "", err
	}

	return processAllImages(ctx, resizeRequest.File, func(img *image.Image) (*image.Image, error) {
		return img.Resize(height, width)
	})
}

func resizePercent(ctx *gin.Context) (string, error) {
	var resizeRequest ResizePercentRequest

	err := ctx.ShouldBind(&resizeRequest)
	if err != nil {
		return "", err
	}

	percent, err := strconv.ParseFloat(resizeRequest.Size, 64)
	if err != nil {
		return "", err
	}

	return processAllImages(ctx, resizeRequest.File, func(img *image.Image) (*image.Image, error) {
		return img.ResizePercent(percent)
	})
}

func grayScaling(ctx *gin.Context) (string, error) {
	var grayScaleRequest GrayScaleRequest

	err := ctx.ShouldBind(&grayScaleRequest)
	if err != nil {
		return "", err
	}

	return processAllImages(ctx, grayScaleRequest.File, func(img *image.Image) (*image.Image, error) {
		return img.GrayScaling()
	})
}

func convertExt(ctx *gin.Context) (string, error) {
	var convertRequest ConvertRequest

	err := ctx.ShouldBind(&convertRequest)
	if err != nil {
		return "", err
	}

	format, err := image.GetFormatFromString(convertRequest.Format)
	if err != nil {
		return "", err
	}

	return processAllImages(ctx, convertRequest.File, func(img *image.Image) (*image.Image, error) {
		return img.Convert(format)
	})
}

func addWatermark(ctx *gin.Context) (string, error) {
	var watermarkRequest WatermarkRequest

	err := ctx.ShouldBind(&watermarkRequest)
	if err != nil {
		return "", err
	}

	return processAllImages(ctx, watermarkRequest.File, func(img *image.Image) (*image.Image, error) {
		return img.Watermark()
	})
}

type processImage func(img *image.Image) (*image.Image, error)

func processAllImages(ctx *gin.Context, file *multipart.FileHeader, processImg processImage) (string, error) {
	imgs, err := unzip(file)
	if err != nil {
		return "", err
	}

	resizedIMGs, err := rangeImages(imgs, processImg)
	if err != nil {
		return "", err
	}

	return zip(ctx, resizedIMGs)
}

func unzip(file *multipart.FileHeader) ([]*image.Image, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}

	return archive.Unzip(f, file.Size)
}

func rangeImages(imgs []*image.Image, process processImage) ([]*image.Image, error) {
	resizedIMGs := make([]*image.Image, len(imgs))
	for i, img := range imgs {
		resizedIMG, err := process(img)
		if err != nil {
			return nil, err
		}

		resizedIMGs[i] = resizedIMG
	}

	return resizedIMGs, nil
}

func zip(ctx *gin.Context, imgs []*image.Image) (string, error) {
	resizedArchive, err := archive.Zip(imgs)
	if err != nil {
		return "", err
	}

	return saveFile(ctx, resizedArchive)
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
