package data

import (
	"testing"
)

// TestUserDatabase verifies that an User can be saved and loaded from the database
func TestUserDatabase(t *testing.T) {
	// Load database configuration
	DB = new(SqliteBackend)
	DB.DSN("~/.config/wavepipe/wavepipe.db")
	if err := DB.Open(); err != nil {
		t.Fatalf("Could not open database connection: %s", err.Error())
	}
	defer DB.Close()

	// Attempt to create and save the user
	user, err := NewUser("TestUser", "TestPassword")
	if err != nil {
		t.Fatalf("Could not create and save user: %s", err.Error())
	}

	// Attempt to load the user
	if err := user.Load(); err != nil {
		t.Fatalf("Could not load user: %s", err.Error())
	}

	// Attempt to delete the user
	if err := user.Delete(); err != nil {
		t.Fatalf("Could not delete user: %s", err.Error())
	}
}

// TestUserJSON verifies that an User can be encoded and decoded from JSON
func TestUserJSON(t *testing.T) {
	// Mock user
	user := new(User)
	user.Username = "TestUser"
	user.Password = "TestPassword"

	// Marshal JSON
	out, err := user.ToJSON()
	if err != nil {
		t.Fatalf("Could not encode JSON: %s", err.Error())
	}

	// Unmarshal
	if err := user.FromJSON(out); err != nil {
		t.Fatalf("Could not decode JSON: %s", err.Error())
	}
}
