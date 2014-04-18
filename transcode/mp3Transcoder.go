package transcode

import (
	"io"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/mdlayher/goset"
)

// MP3Transcoder represents a MP3 transcoding operation
type MP3Transcoder struct {
	Options Options
	ffmpeg  *FFmpeg
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
			Options: MP3CBROptions{
				quality: quality,
			},
		}
	} else {
		// Not an integer, so check for a valid VBR quality
		if !set.New("v0", "V0", "v2", "V2", "v4", "V4").Has(quality) {
			return nil, ErrInvalidQuality
		}

		// Create a VBR transcoder
		transcoder = &MP3Transcoder{
			Options: MP3VBROptions{
				quality: strings.ToUpper(quality),
			},
		}
	}

	// Return configured transcoder
	return transcoder, nil
}

// Codec returns the selected codec used by the transcoder
func (m MP3Transcoder) Codec() string {
	return m.Options.Codec()
}

// Command returns the command invoked by ffmpeg, for debugging
func (m MP3Transcoder) Command() []string {
	// If ffmpeg not started, return no arguments
	if m.ffmpeg == nil {
		return nil
	}

	return append([]string{FFmpegPath}, m.ffmpeg.Arguments()...)
}

// MIMEType returns the MIME type contained within the options
func (m MP3Transcoder) MIMEType() string {
	return m.Options.MIMEType()
}

// Start begins the transcoding process, and returns a stream which contains its output
func (m *MP3Transcoder) Start(song *data.Song) (io.ReadCloser, error) {
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
func (m MP3Transcoder) Quality() string {
	// Check for CBR or VBR
	if _, ok := m.Options.(MP3CBROptions); ok {
		return "CBR " + m.Options.Quality()
	}

	return "VBR " + m.Options.Quality()
}

// Wait waits for the transcoding process to complete, returning an error if it fails
func (m *MP3Transcoder) Wait() error {
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
