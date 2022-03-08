package image

import (
	"bytes"
	"fmt"
	"github.com/sunshineplan/imgconv"
	"image"
	"imgservice/imgerror"
	"os"
	"strings"
	"time"
)

const (
	ErrFormat = imgerror.IMGError("incorrect image format")
)

type Image struct {
	img     image.Image
	format  imgconv.Format
	name    string
	changed time.Time
	bytes   []byte
}

func New(src *bytes.Buffer, info os.FileInfo) (*Image, error) {
	rawImage := src.Bytes()

	img, format, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	imgFmt, err := GetFormatFromString(format)
	if err != nil {
		return nil, err
	}

	return &Image{
		img:     img,
		format:  imgFmt,
		name:    strings.Split(info.Name(), ".")[0],
		changed: info.ModTime(),
		bytes:   rawImage,
	}, nil
}

func (i *Image) Name() string {
	return i.name
}

func (i *Image) Bytes() []byte {
	return i.bytes
}

func (i *Image) FullName() string {
	return fmt.Sprintf("%s.%s", i.name, GetFormat(i.format))
}

func render(img image.Image, name string, format imgconv.Format) (*Image, error) {
	buffer := new(bytes.Buffer)

	err := imgconv.Write(buffer, img, imgconv.FormatOption{Format: format})
	if err != nil {
		return nil, fmt.Errorf("can't render image: %w", err)
	}

	return &Image{
		img:    img,
		format: format,
		name:   strings.Split(name, ".")[0],
		bytes:  buffer.Bytes(),
	}, nil
}

func GetFormat(format imgconv.Format) string {
	switch format {
	case imgconv.JPEG:
		return "jpg"
	case imgconv.PNG:
		return "png"
	case imgconv.GIF:
		return "gif"
	case imgconv.TIFF:
		return "tif"
	case imgconv.BMP:
		return "bmp"
	case imgconv.PDF:
		return "pdf"
	default:
		return "png"
	}
}

func GetFormatFromString(format string) (imgconv.Format, error) {
	switch format {
	case "jpg", "jpeg":
		return imgconv.JPEG, nil
	case "png":
		return imgconv.PNG, nil
	case "gif":
		return imgconv.GIF, nil
	case "tif", "tiff":
		return imgconv.TIFF, nil
	case "bmp":
		return imgconv.BMP, nil
	case "pdf":
		return imgconv.PDF, nil
	default:
		return 0, ErrFormat
	}
}
