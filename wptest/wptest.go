// Package wptest provides common functionality for testing the wavepipe media server.
package wptest

import (
	"math/rand"
	"testing"
	"time"

	"github.com/mdlayher/wavepipe/bindata"
	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/data/models"
)

// MockUser generates a single User with mock data, used for testing.
// The user is randomly generated, but is not guaranteed to be unique.
func MockUser() *models.User {
	return &models.User{
		Username: RandomString(10),
		Password: RandomString(10),
	}
}

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

// WithTemporaryDB generates an in-memory copy of the wavepipe database from
// embedded schema, and is used to provide a disposable copy of the database
// for tests.
func WithTemporaryDB(t *testing.T, fn func(t *testing.T, db *data.DB)) {
	// Retrieve sqlite3 database schema asset
	asset, err := bindata.Asset("res/sqlite3/wavepipe.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Open in-memory database
	wpdb := &data.DB{}
	if err := wpdb.Open("sqlite3", ":memory:"); err != nil {
		t.Fatal(err)
	}

	// Execute schema to build database
	if _, err := wpdb.Exec(string(asset)); err != nil {
		t.Fatal(err)
	}

	// Invoke input closure with test and database
	fn(t, wpdb)

	// Close and destroy database
	if err := wpdb.Close(); err != nil {
		t.Fatal(err)
	}
}
