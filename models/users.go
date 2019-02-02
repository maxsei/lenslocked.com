package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //initializes postgres drivers
	"golang.org/x/crypto/bcrypt"
	"lenslocked.com/hash"
	"lenslocked.com/rand"
)

var (
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound = errors.New("models: resource not found")
	//ErrInvalidID describes when the user enters an invalid ID
	ErrInvalidID = errors.New("models: ID provided was invalid")
	//ErrExistingEmail describes when the user creates account with
	// an email that already exists in the database
	ErrExistingEmail = errors.New("models: email \"%s\" already exists")
	// ErrInvalidEmail describes when the user logs in with an invalid email
	ErrInvalidEmail = errors.New("models: invalid email address provided")
	// ErrInvalidPassword describes	 when the user logs in with an incorrect passwrod
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

const userPwPepper = "nubis"
const hmacSecretKey = "secret-hmac-key"

// NewUserService creates a new connections to the database
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &UserService{
		db:   db,
		hmac: hmac,
	}, nil
}

//UserService details things that can be done with users
type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

// ByID finds a user by a given ID.
// if there is no record to be found nil, and not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail finds a user by a given email.
// if there is no record to be found nil, and not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByRemember will find the given user with the given remember
// token.  This method will handle the hashing  the token for us
// Errors returned will be an internal server error
func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := us.hmac.Hash(token)
	db := us.db.Where("remember_hash = ?", rememberHash)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Authenticate can be used to authenticate a user with a provided username
// and password
func (us *UserService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, ErrInvalidEmail
	}
	foundPasswordBytes := []byte(foundUser.PasswordHash)
	enteredPasswordBytes := []byte(password + userPwPepper)
	err = bcrypt.CompareHashAndPassword(foundPasswordBytes, enteredPasswordBytes)
	switch err {
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	case nil:
		return foundUser, nil
	default:
		return nil, err
	}
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
	pwBytes := []byte(user.Password + userPwPepper)
	hashBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashBytes)
	user.Password = ""

	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}
	user.RememberHash = us.hmac.Hash(user.Remember)

	if err := us.db.Create(user).Error; err != nil {
		return fmt.Errorf(ErrExistingEmail.Error(), user.Email)
	}
	return nil
}

// Update will update the provided user with all of
// the data inside the provided used obect
func (us *UserService) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
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
	Name         string
	Email        string `gorm:"not null;not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gom:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}
