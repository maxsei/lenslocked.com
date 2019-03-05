package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"lenslocked.com/context"
	"lenslocked.com/models"
	"lenslocked.com/rand"
	"lenslocked.com/views"
)

// NewUsers is used to create a new users controller.
// Function will panic if the templates are not parsed
// correctly and should only be used during setup
func NewUsers(us models.UserService) *Users {
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
	us        models.UserService
}

// New renders users templates for the Users type
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, r, nil)
}

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
	var vd views.Data
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.ErrorAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		vd.ErrorAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// LoginForm contains information parsed from the login page
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// Logout form will delete the remember_token in the users browser and
// set the remember token for the user to some new value in the db.
// Finally it redirects the user to the homepage
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	user := context.User(r.Context())
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.Update(user)

	http.Redirect(w, r, "/", http.StatusFound)
}

// Login is used to verify the provided email addres and password
// and then log in the user if they are correct
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		vd.ErrorAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	user, err := u.us.Authenticate(form.Email, form.Password)
	switch err {
	case models.ErrNotFound:
		fmt.Fprintln(w, "Invalid email address.")
	case models.ErrPasswordIncorrect:
		fmt.Fprintln(w, "Invalid password provided.")
	case nil:
		break
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err != nil {
		return
	}
	err = u.signIn(w, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// signIn creates a cookie for the user using their email that expires an hour
// after not refreshing the page.  This function is called for /login and /singup routes
func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}
