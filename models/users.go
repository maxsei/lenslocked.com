package models

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //initializes postgres drivers
	"golang.org/x/crypto/bcrypt"
	"lenslocked.com/hash"
	"lenslocked.com/rand"
)

var (
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound = errors.New("models: resource not found")
	// ErrIDInvalid describes when the user enters an invalid ID
	ErrIDInvalid = errors.New("models: ID provided was invalid")
	// ErrPasswordIncorrect describes	 when the user logs in with an incorrect passwrod
	ErrPasswordIncorrect = errors.New("models: incorrect password provided")
	// ErrEmailRequired describes when an email is not provided
	ErrEmailRequired = errors.New("models: email address is required")
	// ErrEmailInvalid describes when an email does not match a valid email
	ErrEmailInvalid = errors.New("models: email address is not valid ")
	// ErrEmailTaken describes an attempt to create an email that already exists
	ErrEmailTaken = errors.New("models: email address is already taken")
	// ErrPasswordTooShort describes when update or create is attempted with a short password
	ErrPasswordTooShort = errors.New("models: password must be at least eight characters")
	// ErrPasswordRequired describes when a password is not provided
	ErrPasswordRequired = errors.New("models: password is required")
	// ErrRememberTooShort describes when a remember token is not at least 32 bytes
	ErrRememberTooShort = errors.New("models: remember token must be 32 bytes")
	// ErrRememberRequired describes when a remember token is not provided
	ErrRememberRequired = errors.New("models: remember token is required")
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
	uv := newUserValidator(ug, hmac)
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
	// receive ErrNotFound, ErrIDInvalid, or other errors
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
		return nil, ErrPasswordIncorrect
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

// newUserValidator creates a new userValidator with a userDB, hmac,
// and regular expression for emails that need to be matched
func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
}

// ByEmail will normalize the email and then call ByEmail
// on the subsequent UserDB layer
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user,
		uv.normalizeEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

// ByRemember will hash the remember token if necessary
// and then call ByRemember on the subsequent  UserDB layer
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := &User{
		Remember: token,
	}
	if err := runUserValFuncs(user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

// Create will provide a user with a hashed password, discarding
// the entered password.  It will provide the user with a hashed
// remember token and create the user by calling create on the
// subsequent UserDB layer
func (uv *userValidator) Create(user *User) error {
	if err := runUserValFuncs(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.instantiateRemember,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail); err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

// Update will hash the remember token if provided and update
// the user in the subsequent UserDB layer by calling Update
func (uv *userValidator) Update(user *User) error {
	if err := runUserValFuncs(user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.requireEmail,
		uv.normalizeEmail,
		uv.emailFormat); err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

// Delete will check to see if the id of a user trying to be deleted
// is valid.  If it is then Delete is called in the subsequent UserDb
// layer else invalid Id error is return
func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	if err := runUserValFuncs(&user, uv.positiveID); err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

// userValFunc are methods of type userValidator
type userValFunc func(*User) error

// runUserValFuncs runs all the sequential validations methods on the
// given user returning an error if there is any.
func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// bcryptUserPassword takes in a user and bcrypts their password, along
// with some pepper, returning nil for no password to bcrypt and err if
// there was a problem bycripting their password
func (uv *userValidator) bcryptPassword(user *User) error {
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

// hmacRemember hashes the remember token for a given user
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// instantiateRemember creates a new random remember token for a given user
func (uv *userValidator) instantiateRemember(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

// rememberMinBytes returns error if remember token is not base 64 URL
// encoded or less than 32 characters in length.  nil is return if none is provided
func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTooShort
	}
	return nil
}

// rememberHashRequired returns ErrRememberRequired if no hash provided
func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}
	return nil
}

// positiveID returns ErrIDInvalid if ID is non postive
func (uv *userValidator) positiveID(user *User) error {
	if user.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

// requireEmail returns an error if an email is not available
func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

// normalizeEmail trims spaces and makes all letters lowercase
func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

// emailFormat returns an error if email does not math the email regex
func (uv *userValidator) emailFormat(user *User) error {
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// emailIsAvail returns nil if no email is found, err if there is an internal
// error, and ErrEmailTaken if email already exists
func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

// passwordMinLength return nil for no password and ErrPasswordTooShort
// if password is less than eight characters long
func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// passwordRequired returns ErrPasswordRequired if no password provided
func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

// passwordHashRequired returns ErrPasswordRequired if no hash provided
func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

// newUserGorm creates a new connnection to a Gorm db
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
		return ErrIDInvalid
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
