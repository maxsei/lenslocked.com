package controllers

import (
	"fmt"
	"net/http"

	"lenslocked.com/models"
	"lenslocked.com/views"
)

// NewUsers is used to create a new users controller.
// Function will panic if the templates are not parsed
// correctly and should only be used during setup
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

// Users are used to control which template is rendered
// for the templates page.
type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        *models.UserService
}

//
// // New renders the view template for the Users type
// func (u *Users) New(w http.ResponseWriter, r *http.Request) {
// 	if err := u.NewView.Render(w, nil); err != nil {
// 		panic(err)
// 	}
// }

// SignupForm contains information parsed from the signin page
type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Create is used to process the signup form when a user
// submits the sign up form.  This is used to create a new
// user account .
// POST signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)
}

// LoginForm contains information parsed from the login page
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Login is used to verify the provided email addres and password
// and then log in the user if they are correct
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	fmt.Fprintln(w, form)
	// user, err := u.us.ByEmail()
}
