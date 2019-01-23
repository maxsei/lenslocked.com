package views

import (
	"html/template"
)

// View is a wrapper for an html template
type View struct {
	Template *template.Template
	Layout   string
}

// NewView creates an instance of a view.View object
func NewView(layout string, files ...string) *View {
	files = append(files,
		"views/layouts/footer.gohtml",
		"views/layouts/bootstrap.gohtml",
		"views/layouts/navbar.gohtml",
	)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
		Layout:   layout,
	}
}
