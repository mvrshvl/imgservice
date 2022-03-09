package server

import "mime/multipart"

type ResizeRequest struct {
	Height string                `form:"height"`
	Width  string                `form:"width"`
	File   *multipart.FileHeader `form:"file"`
}

type ResizePercentRequest struct {
	Size string                `form:"size"`
	File *multipart.FileHeader `form:"file"`
}

type GrayScaleRequest struct {
	File *multipart.FileHeader `form:"file"`
}

type WatermarkRequest struct {
	File *multipart.FileHeader `form:"file"`
}

type ConvertRequest struct {
	Format string                `form:"format"`
	File   *multipart.FileHeader `form:"file"`
}
