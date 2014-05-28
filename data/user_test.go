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

	// Attempt to update the user
	user.LastFMToken = "hello"
	if err := user.Update(); err != nil {
		t.Fatalf("Could not update user: %s", err.Error())
	}

	// Attempt to delete the user
	if err := user.Delete(); err != nil {
		t.Fatalf("Could not delete user: %s", err.Error())
	}
}
