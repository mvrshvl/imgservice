package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"imgservice/image"
	"io"
)

func Zip(images []*image.Image) ([]byte, error) {
	zipFile := new(bytes.Buffer)

	zipWriter := zip.NewWriter(zipFile)

	for _, img := range images {
		fileInArchive, err := zipWriter.Create(img.FullName())
		if err != nil {
			return nil, err
		}

		rd := bytes.NewReader(img.Bytes())

		_, err = io.Copy(fileInArchive, rd)
		if err != nil {
			return nil, err
		}
	}

	zipWriter.Close()

	return zipFile.Bytes(), nil
}

func Unzip(archive io.ReaderAt, size int64) ([]*image.Image, error) {
	reader, err := zip.NewReader(archive, size)
	if err != nil {
		return nil, err
	}

	var images []*image.Image

	for _, f := range reader.File {
		bufferImg := new(bytes.Buffer)

		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		_, err = bufferImg.ReadFrom(rc)
		if err != nil {
			return nil, err
		}

		err = rc.Close()
		if err != nil {
			return nil, err
		}

		img, err := image.New(bufferImg, f.FileInfo())
		if err != nil {
			fmt.Println(err)

			continue
		}

		images = append(images, img)
	}

	return images, nil
}
