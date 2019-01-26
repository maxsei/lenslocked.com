package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

const (
	layoutDir   string = "views/layouts/"
	templateDir string = "views/"
	templateExt string = ".gohtml"
)

// NewView creates an instance of a view.View object
func NewView(layout string, files ...string) *View {
	formatViewPaths(files)
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

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
}

// Render is used to render a View to the http.ResponseWriter with
// the View Layout and the data provided to fill in template mustaches
func (v View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, nil)
}

// formatViewPath adds the views path to the passed in
// file path strings for views that need to be created as well
// as adding the .gohtml extension to the view strings passed
//
// i.e. "home" outputs "views/home.gohtml"
func formatViewPaths(files []string) {
	for i, f := range files {
		files[i] = templateDir + f + templateExt
	}
}
