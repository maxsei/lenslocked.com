package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if r.URL.Path == "/" {
		fmt.Fprintln(w, "<h1>Welcome to my awesome site!</h1>")
		return
	}
    if r.URL.Path == "/contact" {
        fmt.Fprintf(w, "%s%s\n", "To get in touch, please send an email to ",
            "<a href=\"mailto:support@lenslocked.com\">support@lenslocked.com</a>.")
    }

}

func main() {
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":8080", nil)
}
