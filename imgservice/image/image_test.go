package image

import (
	"bytes"
	"github.com/sunshineplan/imgconv"
	"os"
	"testing"
)

func TestResize(t *testing.T) {
	file, err := os.Open("../test_img.jpeg")
	if err != nil {
		t.Fatal(err)
	}

	src := new(bytes.Buffer)

	_, err = src.ReadFrom(file)
	if err != nil {
		t.Fatal(err)
	}

	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	img, err := New(src, stat)
	if err != nil {
		t.Fatal(err)
	}

	resized, err := img.Resize(100, 100)
	if err != nil {
		t.Fatal(err)
	}

	resizedPercent, err := img.ResizePercent(200)
	if err != nil {
		t.Fatal(err)
	}

	withWatermark, err := img.Watermark()
	if err != nil {
		t.Fatal(err)
	}

	converted, err := img.Convert(imgconv.PDF)
	if err != nil {
		t.Fatal(err)
	}

	gray, err := img.GrayScaling()
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("resized.jpg", resized.Bytes(), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("resizedP.jpg", resizedPercent.Bytes(), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("watermark.jpg", withWatermark.Bytes(), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("converted.pdf", converted.Bytes(), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("gray.jpg", gray.Bytes(), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}
