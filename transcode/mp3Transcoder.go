package transcode

import (
	"strconv"
	"strings"

	"github.com/mdlayher/goset"
)

// MP3Transcoder represents a MP3 transcoding operation
type MP3Transcoder struct {
	cbrQuality string
	vbr        bool
	vbrQuality string
}

// NewMP3Transcoder creates a new MP3 transcoder, and initializes its associated fields
func NewMP3Transcoder(quality string) (*MP3Transcoder, error) {
	// MP3 transcoder instance to return
	transcoder := new(MP3Transcoder)

	// Check if quality is a valid integer, meaning CBR encode
	if cbr, err := strconv.Atoi(quality); err == nil {
		// Check for valid CBR quality
		if !set.New(128, 192, 256, 320).Has(cbr) {
			return nil, ErrInvalidQuality
		}

		// Create a CBR transcoder
		transcoder = &MP3Transcoder{
			cbrQuality: quality,
			vbr:        false,
		}
	} else {
		// Not an integer, so check for a valid VBR quality
		if !set.New("v0", "V0", "v2", "V2", "v4", "V4").Has(quality) {
			return nil, ErrInvalidQuality
		}

		// Create a VBR transcoder
		transcoder = &MP3Transcoder{
			vbr:        true,
			vbrQuality: strings.ToUpper(quality),
		}
	}

	// Return configured transcoder
	return transcoder, nil
}

// Codec returns the name of the codec used by the transcoder
func (m MP3Transcoder) Codec() string {
	return "MP3"
}

// FFmpegCodec returns the name of the codec ffmpeg will use to create the transcode
func (m MP3Transcoder) FFmpegCodec() string {
	return "libmp3lame"
}

// Quality returns the selected quality used by the transcoder
func (m MP3Transcoder) Quality() string {
	// VBR quality
	if m.vbr {
		return "VBR " + m.vbrQuality
	}

	// CBR quality
	return "CBR " + m.cbrQuality
}
