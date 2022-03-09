package archive

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"imgservice/core/image"
)

func TestArchive(t *testing.T) {
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

	img, err := image.New(src, stat)
	if err != nil {
		t.Fatal(err)
	}

	archive, err := Zip([]*image.Image{img})
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("test_archive.zip", archive, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	zipReader := bytes.NewReader(archive)

	imgs, err := Unzip(zipReader, zipReader.Size())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(imgs[0].Bytes(), img.Bytes()) {
		t.Fatalf("bytes is not equal")
	}
}
