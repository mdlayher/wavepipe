package transcode

import (
	"io"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/mdlayher/goset"
)

// OGGTranscoder represents a OGG transcoding operation
type OGGTranscoder struct {
	Options Options
	ffmpeg  *FFmpeg
}

// Codec returns the selected codec used by the transcoder
func (m OGGTranscoder) Codec() string {
	return m.Options.Codec()
}

// Command returns the command invoked by ffmpeg, for debugging
func (m OGGTranscoder) Command() []string {
	// If ffmpeg not started, return no arguments
	if m.ffmpeg == nil {
		return nil
	}

	return append([]string{FFmpegPath}, m.ffmpeg.Arguments()...)
}

// MIMEType returns the MIME type contained within the options
func (m OGGTranscoder) MIMEType() string {
	return m.Options.MIMEType()
}

// Start begins the transcoding process, and returns a stream which contains its output
func (m *OGGTranscoder) Start(song *data.Song) (io.ReadCloser, error) {
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
func (m OGGTranscoder) Quality() string {
	// Check for CBR or VBR
	if _, ok := m.Options.(OGGCBROptions); ok {
		return "CBR " + m.Options.Quality()
	}

	return "VBR " + m.Options.Quality()
}

// Wait waits for the transcoding process to complete, returning an error if it fails
func (m *OGGTranscoder) Wait() error {
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

// cbrSet returns the set of valid CBR qualities for this transcoder
func (m OGGTranscoder) cbrSet() *set.Set {
	return set.New(128, 192, 256, 320, 500)
}

// vbrSet returns the set of valid VBR qualities for this transcoder
func (m OGGTranscoder) vbrSet() *set.Set {
	return set.New("q6", "Q6", "q8", "Q8", "q10", "Q10")
}

// setCBR sets appropriate CBR options for this transcoder
func (m *OGGTranscoder) setCBR(cbr int) {
	m.Options = &OGGCBROptions{strconv.Itoa(cbr)}
}

// setVBR sets appropriate VBR options for this transcoder
func (m *OGGTranscoder) setVBR(vbr string) {
	m.Options = &OGGVBROptions{strings.ToUpper(vbr)}
}
