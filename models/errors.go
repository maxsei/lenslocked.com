package models

import "strings"

const (
	// ErrNotFound is when we cannot find a thing in our database
	ErrNotFound modelError = "models: resource not found"
	// ErrIDInvalid describes when the user enters an invalid ID
	ErrIDInvalid modelError = "models: ID provided was invalid"
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
	// ErrRememberTooShort describes when a remember token is not at least 32 bytes
	ErrRememberTooShort modelError = "models: remember token must be 32 bytes"
	// ErrRememberRequired describes when a remember token is not provided
	ErrRememberRequired modelError = "models: remember token is required"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	return strings.Title(s)
}
