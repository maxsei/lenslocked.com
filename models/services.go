package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //initializes postgres drivers
	"github.com/pkg/errors"
)

type ServicesConfig func(*Services) error

// WithGorm defines a configuration function for services pertaining to
// interaction with a gorm database.  The function requires connection
// information as to which db it connects to as well as the dialect of db
// Supported Dialects:
// -postgres
func WithGorm(dialect, conInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, conInfo)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

// WithUser defines a configuration function for services pertaining to
// CRUD interactions with users in a gorm database. *Requires gorm service
func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		if s.db == nil {
			return ErrNoDBConnection
		}
		s.User = NewUserService(s.db, pepper, hmacKey)
		return nil
	}
}

// WithGallery defines a configuration function for services pertaining to
// CRUD interactions with galleries in a gorm database. *Requires gorm service
func WithGallery() ServicesConfig {
	return func(s *Services) error {
		if s.db == nil {
			return ErrNoDBConnection
		}
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}

// WithImage defines a configuration function for services pertaining to
// CRUD operations on images in the local filesystem.
func WithImage() ServicesConfig {
	return func(s *Services) error {
		if s.db == nil {
			return ErrNoDBConnection
		}
		s.Image = NewImageService()
		return nil
	}
}

// WithLogMode defines a configuration function for toggling LogMode
// on the gorm database
func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		if s.db == nil {
			return ErrNoDBConnection
		}
		s.db.LogMode(mode)
		return nil
	}
}

// NewServices creates a new services struct and returns and error.
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, errors.Wrap(err,
				"make sure your WithGorm Services Config is the first config that is run to establish a db connection")
		}
	}
	return &s, nil
}

// Services contains the type of services this app provides.
type Services struct {
	Gallery GalleryService
	User    UserService
	Image   ImageService
	db      *gorm.DB
}

// Close closes the database connections.
func (s *Services) Close() error {
	return s.db.Close()
}

//DestructiveReset drops all tables and rebuilds them.
func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&User{}, &Gallery{}).Error
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

// AutoMigrate will appempt to automatically migrate all tables.
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{}).Error
}
