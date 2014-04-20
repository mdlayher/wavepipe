package transcode

import (
	"errors"
	"io"

	"github.com/mdlayher/wavepipe/data"

	"github.com/mdlayher/goset"
)

const (
	// FFMpegMP3Codec contains the ffmpeg codec used to transcode to MP3
	FFMpegMP3Codec = "libmp3lame"
	// FFMpegOGGCodec contains the ffmpeg codec used to transcode to Ogg Vorbis
	FFMpegOGGCodec = "libvorbis"
	// OPUSFFmpegCodec contains the ffmpeg codec used to transcode to Opus
	OPUSFFmpegCodec = "libopus"
)

var (
	// ErrInvalidCodec is returned when an invalid transcoder codec is selected
	ErrInvalidCodec = errors.New("transcode: no such transcoder codec")
	// ErrInvalidQuality is returned when an invalid quality is selected for a given codec
	ErrInvalidQuality = errors.New("transcode: invalid quality for transcoder codec")
	// ErrTranscodingDisabled is returned when the transcoding subsystem is disabled, due
	// to not being able to find ffmpeg
	ErrTranscodingDisabled = errors.New("transcode: could not find ffmpeg, transcoding is disabled")
	// ErrMP3Disabled is returned when MP3 transcoding is disabled, due to ffmpeg not
	// containing the necessary codec
	ErrMP3Disabled = errors.New("transcode: "+FFMpegMP3Codec+" codec not found, MP3 transcoding is disabled")
	// ErrOGGDisabled is returned when OGG transcoding is disabled, due to ffmpeg not
	// containing the necessary codec
	ErrOGGDisabled = errors.New("transcode: "+FFMpegOGGCodec+" codec not found, OGG transcoding is disabled")
)

// Enabled determines whether transcoding is available and enabled for wavepipe
var Enabled bool

// FFmpegPath is the path to the ffmpeg binary detected by the transcode manager
var FFmpegPath string

// CodecSet is the set of codecs which wavepipe detected for use with ffmpeg
var CodecSet = set.New()

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
	// Check if transcoding is disabled
	if !Enabled {
		return nil, ErrTranscodingDisabled
	}

	// Check for a valid codec
	switch codec {
	// MP3
	case "MP3":
		// Verify MP3 transcoding is enabled
		if !CodecSet.Has("MP3") {
			return nil, ErrMP3Disabled
		}

		return NewMP3Transcoder(quality)
	// Ogg Vorbis
	case "OGG":
		// Verify OGG transcoding is enabled
		if !CodecSet.Has("OGG") {
			return nil, ErrOGGDisabled
		}

		return NewOGGTranscoder(quality)
	// Invalid choice
	default:
		return nil, ErrInvalidCodec
	}
}
