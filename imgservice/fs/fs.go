package fs

import (
	"github.com/gin-gonic/gin"
	"imgservice/imgerror"
	"os"
)
import "time"
import "net/http"

type InMemoryFS map[string]http.File

const fileNotFound = imgerror.IMGError("file not found")

// Implements FileSystem interface
func (fs InMemoryFS) Open(name string) (http.File, error) {
	f, ok := fs[name]
	if !ok {
		return nil, fileNotFound
	}

	return f, nil
}

func (fs InMemoryFS) Add(name string, data []byte) {
	if name[0] != '/' {
		name = "/" + name
	}

	fs[name] = LoadFile(name, data, fs)
}

type InMemoryFile struct {
	at   int64
	Name string
	data []byte
	fs   InMemoryFS
}

func LoadFile(name string, data []byte, fs InMemoryFS) *InMemoryFile {
	return &InMemoryFile{at: 0,
		Name: name,
		data: data,
		fs:   fs}
}

// Implements the http.File interface
func (f *InMemoryFile) Close() error {
	return nil
}
func (f *InMemoryFile) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{f}, nil
}
func (f *InMemoryFile) Readdir(count int) ([]os.FileInfo, error) {
	res := make([]os.FileInfo, len(f.fs))
	i := 0
	for _, file := range f.fs {
		res[i], _ = file.Stat()
		i++
	}
	return res, nil
}
func (f *InMemoryFile) Read(b []byte) (int, error) {
	i := 0
	for f.at < int64(len(f.data)) && i < len(b) {
		b[i] = f.data[f.at]
		i++
		f.at++
	}
	return i, nil
}
func (f *InMemoryFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.at = offset
	case 1:
		f.at += offset
	case 2:
		f.at = int64(len(f.data)) + offset
	}
	return f.at, nil
}

type InMemoryFileInfo struct {
	file *InMemoryFile
}

// Implements os.FileInfo
func (s *InMemoryFileInfo) Name() string       { return s.file.Name }
func (s *InMemoryFileInfo) Size() int64        { return int64(len(s.file.data)) }
func (s *InMemoryFileInfo) Mode() os.FileMode  { return os.ModeTemporary }
func (s *InMemoryFileInfo) ModTime() time.Time { return time.Time{} }
func (s *InMemoryFileInfo) IsDir() bool        { return false }
func (s *InMemoryFileInfo) Sys() interface{}   { return nil }

const (
	storageKey  = "storage"
	errNotFound = imgerror.IMGError("link unavailable")
	errContext  = imgerror.IMGError("storage is not in context")
	errType     = imgerror.IMGError("incorrect type")
)

func New() InMemoryFS {
	return make(map[string]http.File)
}

func SetCtx(s InMemoryFS) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		ctx.Set(storageKey, s)
	}
}

func GetCtx(ctx *gin.Context) (InMemoryFS, error) {
	v, ok := ctx.Get(storageKey)
	if !ok {
		return nil, errContext
	}

	s, ok := v.(InMemoryFS)
	if !ok {
		return nil, errType
	}

	return s, nil
}
