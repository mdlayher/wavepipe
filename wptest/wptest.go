// Package wptest provides common functionality for testing the wavepipe media server.
package wptest

import (
	"math/rand"
	"time"
)

// RandomString generates a random string of length n.
// Adapted from: http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func RandomString(n int) string {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Random letters slice
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Generate slice of length n
	str := make([]rune, n)
	for i := range str {
		str[i] = letters[rand.Intn(len(letters))]
	}

	return string(str)
}
