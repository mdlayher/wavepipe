package models

// Album represents an album known to wavepipe, and contains information
// extracted from song tags
type Album struct {
	ID     int64
	Artist string
	ArtistID int64
	Title string
	Year  int
}
