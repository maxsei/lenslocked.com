package views

import (
	"lenslocked.com/models"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
	// AlertMsgGeneric is a message we will generally give to users
	// when something goes wrong internally
	AlertMsgGeneric = "Something went Wrong. Please try again later" +
		" and contact us if the problem persists"
)

// Alert will be the data passed to the alert template
type Alert struct {
	Level   string
	Message string
}

// Data is the top level structure that will be passed to our html templates
type Data struct {
	Alert *Alert
	User  *models.User
	Yeild interface{}
}

// ErrorAlert will set the alert type to be generic if it is not an approved Public Error
func (d *Data) ErrorAlert(err error) {
	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: pErr.Public(),
		}
		return
	}
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: AlertMsgGeneric,
	}
}

// SuccessAlert will show an message to show that some action was done successfully
func (d *Data) SuccessAlert(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlSuccess,
		Message: msg,
	}
}

// Public Error defines a type of Error we want to display to the user
type PublicError interface {
	error
	Public() string
}
