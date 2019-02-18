package models

import "github.com/jinzhu/gorm"

// NewServices creates a new services struct and returns and error.
func NewServices(conInfo string) (*Services, error) {
	db, err := gorm.Open("postgres", conInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &Services{
		User:    NewUserService(db),
		Gallery: NewGalleryService(db),
		Image:   NewImageService(),
		db:      db,
	}, nil
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
