package images

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	resizer "github.com/nfnt/resize"
)

type (
	ImageHandler struct {
		dst string
	}

	Service interface {
		Upload(*multipart.FileHeader) (string, error)
		Download(string, int) ([]byte, error)
		GetAllIDs() ([]string, error)
	}
)

var (
	ErrInvalidFormat = errors.New("invalid image format")
	ErrInvalidID     = errors.New("invalid image id")
)

func New(dst string) (Service, error) {
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &ImageHandler{dst: dst}, nil
}

func (f *ImageHandler) GetAllIDs() ([]string, error) {
	var ids []string
	err := filepath.WalkDir(f.dst, func(fPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			tmp := strings.Split(filepath.Base(fPath), ".")
			_, err := uuid.Parse(tmp[0])
			if err != nil {
				return err
			}
			ids = append(ids, tmp[0])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (f *ImageHandler) Download(id string, width int) ([]byte, error) {
	var files []string
	err := filepath.WalkDir(f.dst, func(fPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			id = id + "*"
			if ok, err := path.Match(id, d.Name()); ok && err == nil {
				files = append(files, fPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		return nil, ErrInvalidID
	}

	return resizeImage(files[0], uint(width))
}

func (f *ImageHandler) Upload(file *multipart.FileHeader) (string, error) {
	inFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer inFile.Close()

	buff := make([]byte, 512)
	if _, err = inFile.Read(buff); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buff)
	var extension string
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	default:
		return "", ErrInvalidFormat

	}
	id := uuid.New()
	fullPath := f.dst + "/" + id.String() + extension

	outFile, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return "", nil
	}
	defer outFile.Close()

	// Copy the file to the destination path
	inFile.Seek(0, 0)
	_, err = io.Copy(outFile, inFile)
	if err != nil {
		return "", nil
	}
	return id.String(), nil
}

func resizeImage(path string, width uint) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, imgType, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	//if width = 0, no need to resize
	if width > 0 {
		img = resizer.Resize(width, 0, img, resizer.Lanczos3)
	}

	buffer := new(bytes.Buffer)
	switch imgType {
	case "jpeg":
		err = jpeg.Encode(buffer, img, nil)
	case "png":
		err = png.Encode(buffer, img)
	default:
		return nil, ErrInvalidFormat
	}

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
