package image

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/sunshineplan/imgconv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	maxWidthWatermark  = 240
	maxHeightWatermark = 20

	positionMarkY = 12
	positionMarkX = 0

	maxLenWatermark = 33
)

func (i *Image) Resize(height, width int) (*Image, error) {
	resizedImg := imgconv.Resize(i.img, imgconv.ResizeOption{Height: height, Width: width})

	return render(resizedImg, i.name, i.format)
}

func (i *Image) ResizePercent(percent float64) (*Image, error) {
	resizedImg := imgconv.Resize(i.img, imgconv.ResizeOption{Percent: percent})

	return render(resizedImg, i.name, i.format)
}

func (i *Image) Convert(format imgconv.Format) (*Image, error) {
	return render(i.img, i.name, format)
}

func (i *Image) GrayScaling() (*Image, error) {
	grayImg := image.NewGray(i.img.Bounds())

	for y := i.img.Bounds().Min.Y; y < i.img.Bounds().Max.Y; y++ {
		for x := i.img.Bounds().Min.X; x < i.img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, i.img.At(x, y))
		}
	}

	return render(grayImg, i.name, i.format)
}

func (i *Image) Watermark() (*Image, error) {
	mark := i.generateWatermark()

	img := imgconv.Watermark(i.img, imgconv.WatermarkOption{Mark: mark, Opacity: 200})

	return render(img, i.name, i.format)
}

func (i *Image) generateWatermark() image.Image {
	info := fmt.Sprintf("%s; %s", strings.Split(i.name, ".")[0], i.changed.Format("02 Jan 06 15:04 MST"))

	rectWidth := maxWidthWatermark
	rectHeight := maxHeightWatermark

	if len(info) > maxLenWatermark {
		rectHeight = rectHeight * (len(info)/maxLenWatermark + 1)
	}

	imgWatermark := image.NewRGBA(image.Rect(0, 0, rectWidth, rectHeight))
	for y := imgWatermark.Bounds().Min.Y; y < imgWatermark.Bounds().Max.Y; y++ {
		for x := imgWatermark.Bounds().Min.X; x < imgWatermark.Bounds().Max.X; x++ {
			imgWatermark.Set(x, y, color.RGBA{255, 255, 0, 255})
		}
	}

	var (
		mark      string
		yPosition = positionMarkY
	)

	for j := 0; j < len(info); j += maxLenWatermark {
		if len(info) < j+maxLenWatermark {
			mark = info[j:]
		} else {
			mark = info[j : j+maxLenWatermark]
		}

		d := &font.Drawer{
			Dst:  imgWatermark,
			Src:  image.Black,
			Face: basicfont.Face7x13,
			Dot:  fixed.P(positionMarkX, yPosition),
		}

		d.DrawString(mark)

		yPosition += yPosition
	}

	height, width := i.getWatermarkSize(rectWidth, rectHeight)

	return imgconv.Resize(imgWatermark, imgconv.ResizeOption{Height: height, Width: width})
}

func (i *Image) getWatermarkSize(widthSrc, heightSrc int) (height int, width int) {
	size := i.img.Bounds().Size()

	height = heightSrc * 2
	width = widthSrc * 2

	if size.X < width {
		height = size.X
	}

	if size.Y < height {
		width = size.Y
	}

	return
}
