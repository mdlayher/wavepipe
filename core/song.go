package core

import (
	"errors"
	"os"
	"path"
	"strings"

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
	ID           int
	Album        string
	AlbumID      int `db:"album_id"`
	Artist       string
	ArtistID     int `db:"artist_id"`
	Bitrate      int
	Channels     int
	Comment      string
	FileName     string `db:"file_name"`
	FileSize     int64  `db:"file_size"`
	FileType     string `db:"file_type"`
	Genre        string
	LastModified int64 `db:"last_modified"`
	Length       int
	SampleRate   int `db:"sample_rate"`
	Title        string
	Track        int
	Year         int
}

// SongFromFile creates a new Song from a TagLib file and an os.FileInfo, as created during
// a filesystem walk. Tags and filesystem information are extracted into the struct.
func SongFromFile(file *taglib.File, info os.FileInfo) (*Song, error) {
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

	// Extract file type from the extension, capitalize, drop the dot
	fileType := strings.ToUpper(path.Ext(info.Name()))[1:]

	// Copy over fields from TagLib tags and properties, as well as OS information
	return &Song{
		Album:        file.Album(),
		Artist:       artist,
		Bitrate:      bitrate,
		Channels:     channels,
		Comment:      file.Comment(),
		FileName:     info.Name(),
		FileSize:     info.Size(),
		FileType:     fileType,
		Genre:        file.Genre(),
		LastModified: info.ModTime().Unix(),
		Length:       length,
		SampleRate:   sampleRate,
		Title:        title,
		Track:        file.Track(),
		Year:         file.Year(),
	}, nil
}

// Load pulls an existing Song from the database
func (s *Song) Load() error {
	return db.LoadSong(s)
}

// Save creates a new Song in the database
func (s *Song) Save() error {
	return db.SaveSong(s)
}
