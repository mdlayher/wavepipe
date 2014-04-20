package transcode

import (
	"io"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/mdlayher/goset"
)

// OPUSTranscoder represents a OPUS transcoding operation
type OPUSTranscoder struct {
	Options Options
	ffmpeg  *FFmpeg
}

// NewOPUSTranscoder creates a new OPUS transcoder, and initializes its associated fields
func NewOPUSTranscoder(quality string) (*OPUSTranscoder, error) {
	// OPUS transcoder instance to return
	transcoder := new(OPUSTranscoder)

	// Check if quality is a valid integer, meaning CBR encode
	if cbr, err := strconv.Atoi(quality); err == nil {
		// Check for valid CBR quality
		if !set.New(128, 192, 256, 320, 500).Has(cbr) {
			return nil, ErrInvalidQuality
		}

		// Create a CBR transcoder
		transcoder = &OPUSTranscoder{
			Options: OPUSCBROptions{
				quality: quality,
			},
		}
	} else {
		// Not an integer, so check for a valid VBR quality
		if !set.New("q6", "Q6", "q8", "Q8", "q10", "Q10").Has(quality) {
			return nil, ErrInvalidQuality
		}

		// Create a VBR transcoder
		transcoder = &OPUSTranscoder{
			Options: OPUSVBROptions{
				quality: strings.ToUpper(quality),
			},
		}
	}

	// Return configured transcoder
	return transcoder, nil
}

// Codec returns the selected codec used by the transcoder
func (m OPUSTranscoder) Codec() string {
	return m.Options.Codec()
}

// Command returns the command invoked by ffmpeg, for debugging
func (m OPUSTranscoder) Command() []string {
	// If ffmpeg not started, return no arguments
	if m.ffmpeg == nil {
		return nil
	}

	return append([]string{FFmpegPath}, m.ffmpeg.Arguments()...)
}

// MIMEType returns the MIME type contained within the options
func (m OPUSTranscoder) MIMEType() string {
	return m.Options.MIMEType()
}

// Start begins the transcoding process, and returns a stream which contains its output
func (m *OPUSTranscoder) Start(song *data.Song) (io.ReadCloser, error) {
	// Set up the ffmpeg instance
	m.ffmpeg = NewFFmpeg(song, m.Options)

	// Invoke ffmpeg to create a transcoded audio stream
	if err := m.ffmpeg.Start(); err != nil {
		return nil, err
	}

	// Retrieve the stream from ffmpeg
	return m.ffmpeg.Stream()
}

// Quality returns the selected quality used by the transcoder
func (m OPUSTranscoder) Quality() string {
	// Check for CBR or VBR
	if _, ok := m.Options.(OPUSCBROptions); ok {
		return "CBR " + m.Options.Quality()
	}

	return "VBR " + m.Options.Quality()
}

// Wait waits for the transcoding process to complete, returning an error if it fails
func (m *OPUSTranscoder) Wait() error {
	// Make sure ffmpeg was started, to avoid panic
	if m.ffmpeg == nil {
		return ErrFFmpegNotStarted
	}

	// Wait for ffmpeg
	if err := m.ffmpeg.Wait(); err != nil {
		return err
	}

	// Nullify ffmpeg process
	m.ffmpeg = nil
	return nil
}
