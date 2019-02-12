package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// NewHMAC creates and returns a new HMAC
func NewHMAC(key string) HMAC {
	return HMAC{
		hmac: hmac.New(sha256.New, []byte(key)),
	}
}

// HMAC is a wrapper around the crypto/hmac package
// making it a little easier to use
type HMAC struct {
	hmac hash.Hash
}

// Hash creates a hash of the given input using the private
// key provided when the HMAC was created
func (h HMAC) Hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}
