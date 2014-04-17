package transcode

import (
	"errors"
)

var (
	// ErrInvalidCodec is returned when an invalid transcoder codec is selected
	ErrInvalidCodec = errors.New("transcode: no such transcoder codec")
	// ErrInvalidQuality is returned when an invalid quality is selected for a given codec
	ErrInvalidQuality = errors.New("transcode: invalid quality for transcoder codec")
)

// Enabled determines whether transcoding is available and enabled for wavepipe
var Enabled bool

// FFmpegPath is the path to the ffmpeg binary detected by the transcode manager
var FFmpegPath string

// Transcoder represents a transcoding operation, and the methods which must be defined
// for a transcoder
type Transcoder interface {
	Codec() string
	FFmpegCodec() string
	Quality() string
}

// Factory generates a new Transcoder depending on the input parameters
func Factory(codec string, quality string) (Transcoder, error) {
	// Check for a valid codec
	switch codec {
	// MP3
	case "mp3", "MP3":
		return NewMP3Transcoder(quality)
	// Invalid choice
	default:
		return nil, ErrInvalidCodec
	}
}
