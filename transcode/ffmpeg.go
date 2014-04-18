package transcode

import (
	"github.com/mdlayher/wavepipe/data"
)

// FFmpeg represents the ffmpeg media encoder, and is used to provide a more flexible
// interface than chaining together command-line arguments
type FFmpeg struct {
	song    *data.Song
	options Options
}

// NewFFmpeg creates a new FFmpeg instance using the input song and options
func NewFFmpeg(song *data.Song, options Options) *FFmpeg {
	return &FFmpeg{
		song:    song,
		options: options,
	}
}

// Arguments outputs a slice of the ffmpeg arguments needed to output audio on stdout
func (f FFmpeg) Arguments() []string {
	return []string{
		"-i",
		f.song.FileName,
		"-codec:a",
		f.options.FFmpegCodec(),
		f.options.FFmpegFlags(),
		f.options.FFmpegQuality(),
		"pipe:1." + f.options.Ext(),
	}
}
