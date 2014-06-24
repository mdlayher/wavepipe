package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/mdlayher/wavepipe/data"
)

// TestFactory verifies that the auth.Factory() function is working properly
func TestFactory(t *testing.T) {
	// Table of tests and the expected authenticators
	var tests = []struct {
		path string
		auth AuthenticatorFunc
	}{
		// Root - unauthenticated
		{"/", nilAuthenticate},
		// API root - unauthenticated
		{"/api", nilAuthenticate},
		// API login - bcrypt
		{"/api/v0/login", bcryptAuthenticate},
		// Other API calls - token
		{"/api/v0/status", tokenAuthenticate},
		// Subsonic API - subsonic
		{"/subsonic", subsonicAuthenticate},
		// Bugfix: Last.fm login - token
		{"/api/v0/lastfm/login", tokenAuthenticate},
	}

	// Iterate and verify tests
	for _, test := range tests {
		// Verify proper authenticator chosen via factory
		if auth := Factory(test.path); reflect.TypeOf(auth) != reflect.TypeOf(test.auth) {
			t.Fatalf("mismatched authenticator type: %#v != %#v", auth, test.auth)
		}
	}
}

// TestbasicCredentials verifies that the basicCredentials function is working properly
func Test_basicCredentials(t *testing.T) {
	// Table of tests and expected output
	var tests = []struct {
		header   string
		username string
		password string
		err      error
	}{
		// Empty header
		{"", "", "", ErrEmptyBasic},
		// Missing second element
		{"Basic", "", "", ErrInvalidBasic},
		// Bad header prefix
		{"Digest XXX", "", "", ErrInvalidBasic},
		// Invalid base64
		{"Basic XXX", "", "", ErrInvalidBasic},
		// No colon delimiter
		{"Basic dGVzdHRlc3Q=", "", "", ErrInvalidBasic},
		// Valid credentials
		{"Basic dGVzdDp0ZXN0", "test", "test", nil},
	}

	// Iterate and verify tests
	for _, test := range tests {
		// Fetch credentials
		username, password, err := basicCredentials(test.header)

		// Verify proper username
		if username != test.username {
			t.Fatalf("mismatched username: %v != %v", username, test.username)
		}

		// Verify proper password
		if password != test.password {
			t.Fatalf("mismatched password: %v != %v", password, test.password)
		}

		// Verify proper error
		if err != test.err {
			t.Fatalf("mismatched err: %v != %v", err, test.err)
		}
	}
}

// TestAuthenticate runs through the entire authentication process with a newly-created user
func TestAuthenticate(t *testing.T) {
	// Load database configuration
	data.DB = new(data.SqliteBackend)
	data.DB.DSN("~/.config/wavepipe/wavepipe.db")
	if err := data.DB.Open(); err != nil {
		t.Fatalf("Could not open database connection: %s", err.Error())
	}
	defer data.DB.Close()

	// Create a temporary user, remove it on return
	user, err := data.NewUser("auth_test", "auth_test", data.RoleGuest)
	if err != nil {
		t.Fatal(err)
	}
	defer user.Delete()

	// Table of bcrypt tests and expected output
	var bcryptTests = []struct {
		username  string
		password  string
		clientErr error
	}{
		// No username
		{"", "auth_test", ErrNoUsername},
		// No password
		{"auth_test", "", ErrNoPassword},
		// Invalid user
		{"no_exist", "auth_test", ErrInvalidUsername},
		// Invalid password
		{"auth_test", "bad_pass", ErrInvalidPassword},
		// Correct credentials
		{"auth_test", "auth_test", nil},
	}

	// Iterate all bcrypt tests and check for valid output
	for _, test := range bcryptTests {
		// Generate POST data
		postData := url.Values{}
		postData.Set("username", test.username)
		postData.Set("password", test.password)

		// Generate a HTTP request
		req, err := http.NewRequest("POST", "http://localhost:8080/api/v0/login", bytes.NewBufferString(postData.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		// Set required headers
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Content-Length", strconv.Itoa(len(postData.Encode())))

		// Attempt authentication via bcrypt
		_, authSession, clientErr, serverErr := bcryptAuthenticate(req)

		// Check for nil session, since sessions are not generated on login (they are generated in the API)
		if authSession != nil {
			t.Fatalf("non-nil session: %v", authSession)
		}

		// Check for expected client error
		if clientErr != test.clientErr {
			t.Fatalf("mismatched clientErr: %v != %v", clientErr, test.clientErr)
		}

		// Check for no server errors
		if serverErr != nil {
			t.Fatal(err)
		}
	}

	// Create a temporary session, remove it on return
	session, err := user.CreateSession("auth_test")
	if err != nil {
		t.Fatal(err)
	}
	defer session.Delete()

	// Table of token tests and expected output
	var tokenTests = []struct {
		token     string
		clientErr error
	}{
		// No token
		{"", ErrNoToken},
		// Invalid token
		{"some_token", ErrInvalidToken},
		// Correct token
		{session.Key, nil},
	}

	// Iterate all token tests and check for valid output
	for _, test := range tokenTests {
		// Generate a HTTP request
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/v0/status?s=%s", test.token), nil)
		if err != nil {
			t.Fatal(err)
		}

		// Attempt authentication via token
		_, _, clientErr, serverErr := tokenAuthenticate(req)

		// Check for expected client error
		if clientErr != test.clientErr {
			t.Fatalf("mismatched clientErr: %v != %v", clientErr, test.clientErr)
		}

		// Check for no server errors
		if serverErr != nil {
			t.Fatal(err)
		}
	}
}
