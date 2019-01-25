package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"lenslocked.com/controllers"
	"lenslocked.com/views"
)

var (
	homeView    *views.View
	contactView *views.View
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
	}
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

func main() {
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")
	usersCtrl := controllers.NewUsers()

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/signup", usersCtrl.New)
	http.ListenAndServe(":8080", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func foo() {

}
