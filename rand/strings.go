package rand

import (
	"crypto/rand"
	"encoding/base64"
)

// RememberTokenBytes is the default number of bytes to generate for a token
const RememberTokenBytes = 32

// Bytes will help us generate n random bytes
// or return an error if there was one.
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NBytes returns the number of bytes used in a base64 URL encoding
// and will return an error if the string is not base64 URL encoded
func NBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

// String will help us generate a base 64 encoded string
//  of n random characters and return an error if there is one
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken will help us generate remember
// tokens of a predefined byte size
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}
