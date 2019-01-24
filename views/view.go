package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

const (
	layoutDir   string = "views/layouts/"
	templateExt string = ".gohtml"
)

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

// View is a wrapper for an html template
type View struct {
	Template *template.Template
	Layout   string
}

// Render is used to render a View to the http.ResponseWriter with
// the View Layout and the data provided to fill in template mustaches
func (v View) Render(w http.ResponseWriter, data interface{}) error {
	return v.Template.ExecuteTemplate(w, v.Layout, nil)
}
