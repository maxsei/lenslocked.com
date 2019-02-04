package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //initializes postgres drivers
	"golang.org/x/crypto/bcrypt"
	"lenslocked.com/hash"
	"lenslocked.com/rand"
)

const userPwPepper = "nubis"
const hmacSecretKey = "secret-hmac-key"

var (
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound = errors.New("models: resource not found")
	//ErrInvalidID describes when the user enters an invalid ID
	ErrInvalidID = errors.New("models: ID provided was invalid")
	// ErrInvalidPassword describes	 when the user logs in with an incorrect passwrod
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

// User is a type that represents user model stored in our database
// and is used for user accounts, storying email and password info
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gom:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// NewUserService creates a new connections to the database
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	return &userService{
		UserDB: &userValidator{
			ug,
		},
	}, nil
}

// UserService is a set of methods used to manipulate and
// work with the user model
type UserService interface {
	// Authenticate will verify the provided email address
	// and password are correct. If they are correct, the users
	// correspoding to the email will be returned. Else you will
	// receive ErrNotFound, ErrInvalidID, or other errors
	Authenticate(email, password string) (*User, error)
	UserDB // all methods from UserDB interface
}

// Authenticate can be used to authenticate a user with a provided username
// and password
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
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

// UserDB is used to interact with the users database.
// For pretty much all single user queries:
// if there is no record to be found nil, and error not found is returned
// if there is some other kind of error expect to handle it with 500 error
type UserDB interface {
	// Methods for querying for single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)
	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
	// Used to close a DB connection
	Close() error
	//  Migration helpers
	AutoMigrate() error
	DestructiveReset() error
}

var _ UserService = &userService{}

type userService struct {
	UserDB
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

var _ UserDB = &userGorm{}

// userGorm details things that can be done with databasing users
type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

// ByID finds a user by a given ID.
// if there is no record to be found nil, and error not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail finds a user by a given email.
// if there is no record to be found nil, and error not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// ByRemember finds a user by a given remember token.
// if there is no record to be found nil, and error not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := ug.hmac.Hash(token)
	db := ug.db.Where("remember_hash = ?", rememberHash)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
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
func (ug *userGorm) Create(user *User) error {
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
	user.RememberHash = ug.hmac.Hash(user.Remember)
	return ug.db.Create(user).Error
}

// Update will update the provided user with all of
// the data inside the provided used obect
func (ug *userGorm) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}
	return ug.db.Save(user).Error
}

// Delete will delete a user with a given valid id.
// If the id is invalid Delete will return ErrInvalidID
func (ug *userGorm) Delete(id uint) error {
	if _, err := ug.ByID(id); err == nil {
		ug.db.Delete(User{Model: gorm.Model{ID: id}})
		return nil
	}
	return ErrInvalidID
}

// Close closes the userGorm database connections
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

//DestructiveReset deletes and resets the Users table from the database
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// AutoMigrate will appempt to automatically migrate users table
func (ug *userGorm) AutoMigrate() error {
	return ug.db.AutoMigrate(&User{}).Error
}
