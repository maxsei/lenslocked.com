package views

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
	// AlertMsgGeneric is a message we will generally give to users
	// when something goes wrong internally
	AlertMsgGeneric = "Something went Wrong. Please try again later" +
		"and contact us if the problem persists"
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
