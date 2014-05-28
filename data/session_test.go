package data

import (
	"testing"
)

// TestSessionDatabase verifies that an Session can be saved and loaded from the database
func TestSessionDatabase(t *testing.T) {
	// Load database configuration
	DB = new(SqliteBackend)
	DB.DSN("~/.config/wavepipe/wavepipe.db")
	if err := DB.Open(); err != nil {
		t.Fatalf("Could not open database connection: %s", err.Error())
	}
	defer DB.Close()

	// Attempt to create and save the session
	session, err := NewSession(1, "TestPassword", "TestClient")
	if err != nil {
		t.Fatalf("Could not create and save session: %s", err.Error())
	}

	// Attempt to load the session
	if err := session.Load(); err != nil {
		t.Fatalf("Could not load session: %s", err.Error())
	}

	// Attempt to delete the session
	if err := session.Delete(); err != nil {
		t.Fatalf("Could not delete session: %s", err.Error())
	}
}
