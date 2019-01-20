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
		return
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s%s", "<h1>We could not find the page you were looking for :(</h1>",
		"<p>Please email us if you keep being sent to a invalid page.</p>")
}

func main() {
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":8080", nil)
}
