package transcode

import (
	"errors"
	"io"
	"os/exec"

	"github.com/mdlayher/wavepipe/data"
)

var (
	ErrFFmpegNotStarted = errors.New("ffmpeg: transcoding process has not started")
)

// FFmpeg represents the ffmpeg media encoder, and is used to provide a more flexible
// interface than chaining together command-line arguments
type FFmpeg struct {
	ffmpeg  *exec.Cmd
	options Options
	song    *data.Song
	started bool
	stream  io.ReadCloser
}

// NewFFmpeg creates a new FFmpeg instance using the input song and options
func NewFFmpeg(song *data.Song, options Options) *FFmpeg {
	return &FFmpeg{
		options: options,
		song:    song,
		started: false,
	}
}

// Arguments outputs a slice of the ffmpeg arguments needed to output audio on stdout
func (f FFmpeg) Arguments() []string {
	return []string{
		"-i",
		f.song.FileName,
		"-acodec",
		f.options.FFmpegCodec(),
		f.options.FFmpegFlags(),
		f.options.FFmpegQuality(),
		"pipe:1." + f.options.Ext(),
	}
}

// Start invokes the ffmpeg media encoder using the path discovered by the transcode manager
func (f *FFmpeg) Start() error {
	// Generate the ffmpeg instance
	f.ffmpeg = exec.Command(FFmpegPath, f.Arguments()...)

	// Generate the output stream
	stream, err := f.ffmpeg.StdoutPipe()
	if err != nil {
		return err
	}
	f.stream = stream

	// Invoke the process
	if err := f.ffmpeg.Start(); err != nil {
		return err
	}

	// Mark ffmpeg started
	f.started = true
	return nil
}

// Stream returns the current stream which ffmpeg is feeding while started
func (f FFmpeg) Stream() (io.ReadCloser, error) {
	// Verify ffmpeg is running
	if !f.started {
		return nil, ErrFFmpegNotStarted
	}

	// Return stream
	return f.stream, nil
}

// Wait waits for the ffmpeg instance to exit
func (f *FFmpeg) Wait() error {
	// Verify ffmpeg is running
	if !f.started {
		return ErrFFmpegNotStarted
	}

	// Wait for exit
	if err := f.ffmpeg.Wait(); err != nil {
		return err
	}

	// Stopped!
	f.started = false
	return nil
}
