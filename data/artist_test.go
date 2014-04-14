package data

import (
	"testing"
)

// Mock artist
var artist = Artist{
	Title: "TestArtist",
}

// TestArtistDatabase verifies that an Artist can be saved and loaded from the database
func TestArtistDatabase(t *testing.T) {
	// Load database configuration
	DB = new(SqliteBackend)
	DB.DSN("~/.config/wavepipe/wavepipe.db")

	// Attempt to save the artist
	if err := artist.Save(); err != nil {
		t.Fatalf("Could not save artist: %s", err.Error())
	}

	// Attempt to load the artist
	if err := artist.Load(); err != nil {
		t.Fatalf("Could not load artist: %s", err.Error())
	}

	// Attempt to delete the artist
	if err := artist.Delete(); err != nil {
		t.Fatalf("Could not delete artist: %s", err.Error())
	}
}

// TestArtistJSON verifies that an Artist can be encoded and decoded from JSON
func TestArtistJSON(t *testing.T) {
	// Marshal JSON
	out, err := artist.ToJSON()
	if err != nil {
		t.Fatalf("Could not encode JSON: %s", err.Error())
	}

	// Unmarshal
	if err := artist.FromJSON(out); err != nil {
		t.Fatalf("Could not decode JSON: %s", err.Error())
	}
}
