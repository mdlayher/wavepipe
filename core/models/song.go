package models

import (
	"errors"

	"github.com/wtolson/go-taglib"
)

var (
	// ErrSongTags is returned when required tags could not be extracted from a TagLib file
	ErrSongTags = errors.New("song: required tags could be extracted from TagLib file")
	// ErrSongProperties is returned when required properties could not be extracted from a TagLib file
	ErrSongProperties = errors.New("song: required properties could be extracted from TagLib file")
)

// Song represents a song known to wavepipe, and contains metadata regarding
// the song, and where it resides in the filsystem
type Song struct {
	ID           int64  `beedb:"PK" sql:"id" tname:"songs"`
	Album        string `sql:"-"`
	AlbumID      int64  `sql:"album_id"`
	Artist       string `sql:"-"`
	ArtistID     int64  `sql:"artist_id"`
	Bitrate      int    `sql:"bitrate"`
	Channels     int    `sql:"channels"`
	Comment      string `sql:"comment"`
	FileName     string `sql:"file_name"`
	FileSize     int64  `sql:"file_size"`
	Genre        string `sql:"genre"`
	LastModified int64  `sql:"last_modified"`
	Length       int    `sql:"length"`
	SampleRate   int    `sql:"sample_rate"`
	Title        string `sql:"title"`
	Track        int    `sql:"track"`
	Year         int    `sql:"year"`
}

// SongFromFile creates a new Song from a TagLib file, extracting its tags
// and properties to build the struct
func SongFromFile(file *taglib.File) (*Song, error) {
	// Retrieve some tags needed by wavepipe, check for empty
	// At minimum, we will need an artist and title to do anything useful with this file
	title := file.Title()
	artist := file.Artist()
	if title == "" || artist == "" {
		return nil, ErrSongTags
	}

	// Retrieve all properties, check for empty
	// Note: length will probably be more useful as an integer, and a Duration method can
	// be added later on if needed
	bitrate := file.Bitrate()
	channels := file.Channels()
	length := int(file.Length().Seconds())
	sampleRate := file.Samplerate()

	if bitrate == 0 || channels == 0 || length == 0 || sampleRate == 0 {
		return nil, ErrSongProperties
	}

	// Copy over fields from TagLib tags and properties
	return &Song{
		Album:      file.Album(),
		Artist:     artist,
		Bitrate:    bitrate,
		Channels:   channels,
		Comment:    file.Comment(),
		Genre:      file.Genre(),
		Length:     length,
		SampleRate: sampleRate,
		Title:      title,
		Track:      file.Track(),
		Year:       file.Year(),
	}, nil
}
