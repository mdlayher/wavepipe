package data

import (
	"io"
	"os"
)

// Art represents folder or album art known to wavepipe, and contains filesystem metadata
type Art struct {
	ID           int
	FileSize     int64  `db:"file_size"`
	FileName     string `db:"file_name"`
	LastModified int64  `db:"last_modified"`
}

// Delete removes existing Art from the database
func (a *Art) Delete() error {
	return DB.DeleteArt(a)
}

// Load pulls existing Art from the database
func (a *Art) Load() error {
	return DB.LoadArt(a)
}

// Save creates new Art in the database
func (a *Art) Save() error {
	return DB.SaveArt(a)
}

// Update updates existing Art in the database
func (a *Art) Update() error {
	return DB.UpdateArt(a)
}

// Stream returns an art stream from the art file
func (a Art) Stream() (io.ReadCloser, error) {
	return os.Open(a.FileName)
}
