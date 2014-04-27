package data

import (
	"testing"
)

// Mock folder
var folder = Folder{
	ParentID: 1,
	Title:    "TestFolder",
	Path:     "/some/folder",
}

// TestFolderDatabase verifies that an Folder can be saved and loaded from the database
func TestFolderDatabase(t *testing.T) {
	// Load database configuration
	DB = new(SqliteBackend)
	DB.DSN("~/.config/wavepipe/wavepipe.db")
	if err := DB.Open(); err != nil {
		t.Fatalf("Could not open database connection: %s", err.Error())
	}
	defer DB.Close()

	// Attempt to save the folder
	if err := folder.Save(); err != nil {
		t.Fatalf("Could not save folder: %s", err.Error())
	}

	// Attempt to load the folder
	if err := folder.Load(); err != nil {
		t.Fatalf("Could not load folder: %s", err.Error())
	}

	// Attempt to delete the folder
	if err := folder.Delete(); err != nil {
		t.Fatalf("Could not delete folder: %s", err.Error())
	}
}

// TestFolderJSON verifies that an Folder can be encoded and decoded from JSON
func TestFolderJSON(t *testing.T) {
	// Marshal JSON
	out, err := folder.ToJSON()
	if err != nil {
		t.Fatalf("Could not encode JSON: %s", err.Error())
	}

	// Unmarshal
	if err := folder.FromJSON(out); err != nil {
		t.Fatalf("Could not decode JSON: %s", err.Error())
	}
}
