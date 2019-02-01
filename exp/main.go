package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"lenslocked.com/hash"
)

func main() {
	toHash := []byte("thisismystringtohash")
	h := hmac.New(sha256.New, []byte("secretkey"))
	h.Reset()
	h.Write(toHash)
	b := h.Sum(nil)
	fmt.Println(base64.URLEncoding.EncodeToString(b))

	hmac := hash.NewHMAC("secretkey")
	fmt.Println(hmac.Hash("thisismystringtohash"))
}
