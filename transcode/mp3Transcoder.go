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

// cbrSet returns the set of valid CBR qualities for this transcoder
func (m MP3Transcoder) cbrSet() *set.Set {
	return set.New(128, 192, 256, 320)
}

// vbrSet returns the set of valid VBR qualities for this transcoder
func (m MP3Transcoder) vbrSet() *set.Set {
	return set.New("v0", "V0", "v2", "V2", "v4", "V4")
}

// setCBR sets appropriate CBR options for this transcoder
func (m *MP3Transcoder) setCBR(cbr int) {
	m.Options = &MP3CBROptions{strconv.Itoa(cbr)}
}

// setVBR sets appropriate VBR options for this transcoder
func (m *MP3Transcoder) setVBR(vbr string) {
	m.Options = &MP3VBROptions{strings.ToUpper(vbr)}
}
