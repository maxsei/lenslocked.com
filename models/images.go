package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]string, error)
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {
	defer r.Close()
	path, err := is.mkImagePath(galleryID)
	if err != nil {
		return err
	}
	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *imageService) ByGalleryID(galleryID uint) ([]string, error) {
	path := is.imagePath(galleryID)
	paths, err := filepath.Glob(path + "*")
	for i := range paths {
		paths[i] = "/" + paths[i]
		paths[i] = strings.Replace(paths[i], "\\", "/", -1)
	}
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := is.imagePath(galleryID)
	if err := os.MkdirAll(galleryPath, 0755); err != nil {
		return "", err
	}
	return galleryPath, nil
}
