package transcode

import (
	"errors"
	"io"

	"github.com/mdlayher/wavepipe/data"
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
	Command() []string
	MIMEType() string
	Start(*data.Song) (io.ReadCloser, error)
	Wait() error
	Quality() string
}

// Options represents an audio codec and its quality settings, and includes methods to
// retrieve these settings
type Options interface {
	Codec() string
	Ext() string
	FFmpegCodec() string
	FFmpegFlags() string
	FFmpegQuality() string
	MIMEType() string
	Quality() string
}

// Factory generates a new Transcoder depending on the input parameters
func Factory(codec string, quality string) (Transcoder, error) {
	// Check for a valid codec
	switch codec {
	// MP3
	case "mp3", "MP3":
		return NewMP3Transcoder(quality)
	// Ogg Vorbis
	case "ogg", "OGG":
		return NewOGGTranscoder(quality)
	// Invalid choice
	default:
		return nil, ErrInvalidCodec
	}
}
