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
		User: NewUserService(db),
	}, nil
}

// Services contains the type of services this app provides.
type Services struct {
	Gallery GalleryService
	User    UserService
}
