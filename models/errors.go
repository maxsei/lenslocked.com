package models

import "strings"

const (
	// ErrNoDBConnection is when no database connection is established
	ErrNoDBConnection modelError = "models : no db connection found when required"
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound modelError = "models: resource not found"
	// ErrPasswordIncorrect describes	 when the user logs in with an incorrect passwrod
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	// ErrEmailRequired describes when an email is not provided
	ErrEmailRequired modelError = "models: email address is required"
	// ErrEmailInvalid describes when an email does not match a valid email
	ErrEmailInvalid modelError = "models: email address is not valid "
	// ErrEmailTaken describes an attempt to create an email that already exists
	ErrEmailTaken modelError = "models: email address is already taken"
	// ErrPasswordTooShort describes when update or create is attempted with a short password
	ErrPasswordTooShort modelError = "models: password must be at least eight characters"
	// ErrPasswordRequired describes when a password is not provided
	ErrPasswordRequired modelError = "models: password is required"
	// ErrTitleRequired describes when a gallery title is not provided on the galleries page
	ErrTitleRequired modelError = "models: gallery title is required"
	// ErrRememberTooShort describes when a remember token is not at least 32 bytes
	ErrRememberTooShort privateError = "models: remember token must be 32 bytes"
	// ErrRememberRequired describes when a remember token is not provided
	ErrRememberRequired privateError = "models: remember token is required"
	// ErrIDInvalid describes when the user enters an invalid ID
	ErrIDInvalid privateError = "models: ID provided was invalid"
	// ErrUserIDRequired describes when a user ID is not provided on the galleries page
	ErrUserIDRequired privateError = "models: user ID is required"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	return strings.Title(s)
}

type privateError string

func (e privateError) Error() string {
	return string(e)
}
