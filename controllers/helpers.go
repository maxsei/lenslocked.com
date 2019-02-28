package controllers

import (
	"net/http"
	"path/filepath"

	schema "github.com/gorilla/Schema"
)

var (
	// PermittedExtensions are the file types that are allowed for uploads
	PermittedExtensions = []string{"jpg", "jpeg", "png"}
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(dst, r.PostForm); err != nil {
		return err
	}
	return nil
}

func hasPermittedExension(filename string) bool {
	extension := filepath.Ext(filename)
	for _, ext := range PermittedExtensions {
		if ext == extension {
			return true
		}
	}
	return false
}

func invalidExension(filename string) bool {
	return !hasPermittedExension(filename)
}
