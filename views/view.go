package views

import (
	"html/template"
	"path/filepath"
)

const (
	layoutDir   string = "views/layouts/"
	templateExt string = ".gohtml"
)

// View is a wrapper for an html template
type View struct {
	Template *template.Template
	Layout   string
}

// NewView creates an instance of a view.View object
func NewView(layout string, files ...string) *View {
	layouts, err := filepath.Glob(layoutDir + "*" + templateExt)
	if err != nil {
		panic(err)
	}
	files = append(files, layouts...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
		Layout:   layout,
	}
}
