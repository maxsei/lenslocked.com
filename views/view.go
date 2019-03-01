package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"lenslocked.com/context"
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
	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField is not yet implemented")
		},
	}).ParseFiles(files...)
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
	v.Render(w, r, nil)
}

// Render is used to render a View to the http.ResponseWriter with
// the View Layout and the data provided to fill in template mustaches
// if no data is not of type views.Data then create a new one with yeild data
// as the data passedinto the Render method
func (v View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	fmt.Println(r.URL.Path)
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{
			Yeild: data,
		}
	}
	csrfFeild := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfFeild
		},
	})
	vd.User = context.User(r.Context())
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		log.Println(err)
		http.Error(w, "Oops something went wrong.", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
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
