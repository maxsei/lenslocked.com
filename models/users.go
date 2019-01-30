package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //initializes postgres drivers
)

var (
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound = errors.New("models: resource not found")
)

//UserService details things that can be done with users
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new connections to the database
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

// ByID finds a user by a given ID
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	err := us.db.Where("id = ?", id).First(&user).Error
	switch err {
	case nil:
		return &user, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Create will create the provied user and back fill
// the Id, createdAt, and Updataed At
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Close closes the UserService database connections
func (us *UserService) Close() error {
	return us.db.Close()
}

//DestructiveReset deletes and resets the Users table from the database
func (us *UserService) DestructiveReset() {
	us.db.DropTableIfExists(&User{})
	us.db.AutoMigrate(&User{})
}

// User is a type that describes users on the site
type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;not null;unique_index"`
}
