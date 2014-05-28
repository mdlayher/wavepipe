package data

// Artist represents an artist known to wavepipe, and contains a unique ID
// and name for this artist
type Artist struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ArtistFromSong creates a new Artist from a Song model, extracting its
// fields as needed to build the struct
func ArtistFromSong(song *Song) *Artist {
	// Copy the artist name to title
	return &Artist{
		Title: song.Artist,
	}
}

// Delete removes an existing Artist from the database
func (a *Artist) Delete() error {
	return DB.DeleteArtist(a)
}

// Load pulls an existing Artist from the database
func (a *Artist) Load() error {
	return DB.LoadArtist(a)
}

// Save creates a new Artist in the database
func (a *Artist) Save() error {
	return DB.SaveArtist(a)
}
