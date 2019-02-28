package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Image is not stored in the db
type Image struct {
	GalleryID uint
	Filename  string
}

// RelPath returns the relative filepath to the associated image in the file system
func (i *Image) RelPath() string {
	return "/" + i.RootPath()
}

// RootPath returns the path starting at the root of the lenslocked project
// to the associated image in the file system
func (i *Image) RootPath() string {
	return fmt.Sprintf("images/galleries/%d/%s", i.GalleryID, i.Filename)
}

type ImageService interface {
	Create(img *Image, r io.ReadCloser) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(img *Image) error
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct{}

func (is *imageService) Create(img *Image, r io.ReadCloser) error {
	defer r.Close()
	path, err := is.mkImagePath(img.GalleryID)
	if err != nil {
		return err
	}
	dst, err := os.Create(path + img.Filename)
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

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imagePath(galleryID)
	paths, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	images := make([]Image, len(paths))
	for i := range paths {
		images[i].Filename = filepath.Base(paths[i])
		images[i].GalleryID = galleryID
	}
	return images, nil
}
func (is *imageService) Delete(img *Image) error {
	return os.Remove(img.RootPath())
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
