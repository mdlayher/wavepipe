package core

// Album represents an album known to wavepipe, and contains information
// extracted from song tags
type Album struct {
	ID       int
	Artist   string
	ArtistID int `db:"artist_id"`
	Title    string
	Year     int
}

// AlbumFromSong creates a new Album from a Song model, extracting its
// fields as needed to build the struct
func AlbumFromSong(song *Song) *Album {
	return &Album{
		Artist: song.Artist,
		Title:  song.Album,
		Year:   song.Year,
	}
}

// Load pulls an existing Album from the database
func (a *Album) Load() error {
	return db.LoadAlbum(a)
}

// Save creates a new Album in the database
func (a *Album) Save() error {
	return db.SaveAlbum(a)
}
