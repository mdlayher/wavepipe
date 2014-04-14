package data

import (
	"testing"
)

// Mock song
var song = Song{
	AlbumID:  1,
	ArtistID: 1,
	FileName: "/mem/test",
	Artist:   "TestArtist",
	Album:    "TestAlbum",
	Title:    "TestSong",
	Year:     2014,
}

// TestSongDatabase verifies that an Song can be saved and loaded from the database
func TestSongDatabase(t *testing.T) {
	// Load database configuration
	DB = new(SqliteBackend)
	DB.DSN("~/.config/wavepipe/wavepipe.db")

	// Attempt to save the song
	if err := song.Save(); err != nil {
		t.Fatalf("Could not save song: %s", err.Error())
	}

	// Attempt to load the song
	if err := song.Load(); err != nil {
		t.Fatalf("Could not load song: %s", err.Error())
	}

	// Attempt to delete the song
	if err := song.Delete(); err != nil {
		t.Fatalf("Could not delete song: %s", err.Error())
	}
}

// TestSongJSON verifies that an Song can be encoded and decoded from JSON
func TestSongJSON(t *testing.T) {
	// Marshal JSON
	out, err := song.ToJSON()
	if err != nil {
		t.Fatalf("Could not encode JSON: %s", err.Error())
	}

	// Unmarshal
	if err := song.FromJSON(out); err != nil {
		t.Fatalf("Could not decode JSON: %s", err.Error())
	}
}
