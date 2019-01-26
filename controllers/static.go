package controllers

import (
	"lenslocked.com/views"
)

// NewStatic returns a new static structure
func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "static/home"),
		Contact: views.NewView("bootstrap", "static/contact"),
	}
}

// Static contains all the views used in the site
type Static struct {
	Home    *views.View
	Contact *views.View
}
