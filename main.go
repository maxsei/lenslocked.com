package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maxsei/lenslocked.com/views"
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
	fmt.Println("Listening and serving on localhost:8080")

	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	http.ListenAndServe(":8080", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
