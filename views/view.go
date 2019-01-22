package views

import (
	"html/template"
)

// View is a wrapper for an html template
type View struct {
	Template *template.Template
}

// NewView creates an instance of a view.View object
func NewView(files ...string) *View {
	files = append(files, "views/layouts/footer.gohtml")
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
	}
}
