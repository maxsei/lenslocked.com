package models

import (
	"errors"

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
	// ErrInvalidPassword describes	 when the user logs in with an incorrect passwrod
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

const userPwPepper = "nubis"
const hmacSecretKey = "secret-hmac-key"

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
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := &userValidator{
		UserDB: ug,
		hmac:   hmac,
	}
	return &userService{
		UserDB: uv,
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

type userValFunc func(*User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac hash.HMAC
}

// ByRemember will hash the remember token if necessary
// and then call ByRemember on the subsequent  UserDB layer
func (uv *userValidator) ByRemember(token string) (*User, error) {
	rememberHash := uv.hmac.Hash(token)
	return uv.UserDB.ByRemember(rememberHash)
}

// Create will provide a user with a hashed password, discarding
// the entered password.  It will provide the user with a hashed
// remember token and create the user by calling create on the
// subsequent UserDB layer
func (uv *userValidator) Create(user *User) error {
	if err := runUserValFuncs(user, uv.bcryptUserPassword); err != nil {
		return err
	}
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return uv.UserDB.Create(user)
}

// Update will hash the remember token if provided and update
// the user in the subsequent UserDB layer by calling Update
func (uv *userValidator) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = uv.hmac.Hash(user.Remember)
	}
	return uv.UserDB.Update(user)
}

// Delete will check to see if the id of a user trying to be deleted
// is valid.  If it is then Delete is called in the subsequent UserDb
// layer else invalid Id error is return
func (uv *userValidator) Delete(id uint) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return uv.UserDB.Delete(id)
}

// bcryptUserPassword takes in a user and bcrypts their password, along
// with some pepper, returning nil for no password to bcrypt and err if
// there was a problem bycripting their password
func (uv *userValidator) bcryptUserPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPwPepper)
	hashBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashBytes)
	user.Password = ""
	return nil
}
func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &userGorm{
		db: db,
	}, nil
}

var _ UserDB = &userGorm{}

// userGorm details things that can be done with databasing users
type userGorm struct {
	db *gorm.DB
}

// ByID finds a user by a given ID.
// if there is no record to be found nil, and error not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
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

// ByRemember finds a user by a given already hashed remember token.
// if there is no record to be found nil, and error not found is returned
// if there is some other kind of error expect to handle it with 500 error
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	db := ug.db.Where("remember_hash = ?", rememberHash)
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

// Create insert a given user in the Gorm db.
// Errors returned should be handled with a 500 StatusInternalServerError
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update will update new data for a user in the Gorm db.
// Errors returned should be handled with a 500 StatusInternalServerError
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete will delete a user with a given id.
// If the id is invalid it will delete the entire db
func (ug *userGorm) Delete(id uint) error {
	if id <= 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
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
