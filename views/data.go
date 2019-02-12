package views

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
	Yeild interface{}
}

// SetAlert will set the alert type to be generic if it is not an approved Public Error
func (d *Data) SetAlert(err error) {
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

// Public Error defines a type of Error we want to display to the user
type PublicError interface {
	error
	Public() string
}
