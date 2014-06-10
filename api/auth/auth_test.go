package auth

import (
	"reflect"
	"testing"
)

// TestFactory verifies that the auth.Factory() function is working properly
func TestFactory(t *testing.T) {
	// Table of tests and the expected authenticators
	var tests = []struct {
		path string
		auth AuthMethod
	}{
		// Root - unauthenticated
		{"/", nil},
		// API root - unauthenticated
		{"/api", nil},
		// API login - bcrypt
		{"/api/v0/login", new(BcryptAuth)},
		// Other API calls - token
		{"/api/v0/status", new(TokenAuth)},
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
			t.Fatalf("mistmatched username: %v != %v", username, test.username)
		}

		// Verify proper password
		if password != test.password {
			t.Fatalf("mistmatched password: %v != %v", password, test.password)
		}

		// Verify proper error
		if err != test.err {
			t.Fatalf("mistmatched err: %v != %v", err, test.err)
		}
	}
}
