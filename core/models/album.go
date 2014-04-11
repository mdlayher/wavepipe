package models

// Album represents an album known to wavepipe, and contains information
// extracted from song tags
type Album struct {
	ID       int64  `beedb:"PK" sql:"id" tname:"albums"`
	Artist   string `sql:"-"`
	ArtistID int64  `sql:"artist_id"`
	Title    string `sql:"title"`
	Year     int    `sql:"year"`
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
