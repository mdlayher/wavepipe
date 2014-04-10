package models

import (
	"errors"

	"github.com/vbatts/go-taglib/taglib"
)

var (
	// ErrSongTags is returned when the tags could not be extracted from a TagLib file
	ErrSongTags = errors.New("song: no tags could be extracted from TagLib file")
	// ErrSongProperties is returned when the properties could not be extracted from a TagLib file
	ErrSongProperties = errors.New("song: no properties could be extracted from TagLib file")
)

// Song represents a song known to wavepipe, and contains metadata regarding
// the song, and where it resides in the filsystem
type Song struct {
	Album      string
	Artist     string
	Bitrate    int
	Channels   int
	Comment    string
	Genre      string
	Length     int
	SampleRate int
	Track      int
	Year       int
}

// SongFromFile creates a new Song from a TagLib file, extracting its tags
// and properties to build the struct
func SongFromFile(file *taglib.File) (*Song, error) {
	// Retrieve tags, check for empty
	tags := file.GetTags()
	if tags.Title == "" && tags.Artist == "" && tags.Album == "" {
		return nil, ErrSongTags
	}

	// Retrieve properties, check for empty
	properties := file.GetProperties()
	if properties.Bitrate == 0 && properties.Channels == 0 && properties.Length == 0 {
		return nil, ErrSongProperties
	}

	// Copy over fields from TagLib tags and properties
	return &Song{
		Album:      tags.Album,
		Artist:     tags.Artist,
		Bitrate:    properties.Bitrate,
		Channels:   properties.Channels,
		Comment:    tags.Comment,
		Genre:      tags.Genre,
		Length:     properties.Length,
		SampleRate: properties.Samplerate,
		Track:      tags.Track,
		Year:       tags.Year,
	}, nil
}
