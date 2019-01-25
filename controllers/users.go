package controllers

import (
	"net/http"

	"lenslocked.com/views"
)

// NewUsers is used to create a new users controller.
// Function will panic if the templates are not parsed
// correctly and should only be used during setup
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

// Users are used to control which template is rendered
// for the templates page.
type Users struct {
	NewView *views.View
}

// New renders the view template for the Users type
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}
