package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //initializes postgres drivers
)

var (
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound = errors.New("models: resource not found")
	//ErrInvalidID describes when the user enters an invalid ID
	ErrInvalidID = errors.New("models: ID provided was invalid")
	//ErrDuplicateKey describes when the user enters an duplicate key
	ErrDuplicateKey = errors.New("models: Unique key \"%s\" already exists")
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
	// db.AutoMigrate(&User{})  do this if something goes wrong in testing
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

// ByID finds a user by a given ID.
// if there is no record to be found nil, and not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// ByEmail finds a user by a given email.
// if there is no record to be found nil, and not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// first will query using the provided gorm.DB and
// will get the first item returned and place it into
// dst.  It will return error not found if nothing is found
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// Create will create the provied user and back fill
// the Id, createdAt, and Updataed At
func (us *UserService) Create(user *User) error {
	if err := us.db.Create(user).Error; err != nil {
		return fmt.Errorf(ErrDuplicateKey.Error(), user.Email)
	}
	return nil
}

// Update will update the provided user with all of
// the data inside the provided used obect
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Delete will delete a user with a given valid id.
// If the id is invalid Delete will return ErrInvalidID
func (us *UserService) Delete(id uint) error {
	if _, err := us.ByID(id); err == nil {
		us.db.Delete(User{Model: gorm.Model{ID: id}})
		return nil
	}
	return ErrInvalidID
}

// Close closes the UserService database connections
func (us *UserService) Close() error {
	return us.db.Close()
}

//DestructiveReset deletes and resets the Users table from the database
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

// AutoMigrate will appempt to automatically migrate users table
func (us *UserService) AutoMigrate() error {
	return us.db.AutoMigrate(&User{}).Error
}

// User is a type that describes users on the site
type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;not null;unique_index"`
}
