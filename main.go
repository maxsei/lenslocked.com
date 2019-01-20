package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "<h1>Welcome to my awesome site!</h1>")

}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "%s%s\n", "To get in touch, please send an email to ",
		"<a href=\"mailto:support@lenslocked.com\">support@lenslocked.com</a>.")
}

func main() {
	fmt.Println("saved creds")
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	http.ListenAndServe(":8080", r)
}
