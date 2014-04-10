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

// AlbumFromSong creates a new Album from a Song model, extracting its
// fields as needed to build the struct
func AlbumFromSong(song *Song) *Album {
	return &Album{
		Artist: song.Artist,
		Title:  song.Album,
		Year:   song.Year,
	}
}
