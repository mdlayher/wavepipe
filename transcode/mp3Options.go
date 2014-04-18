package transcode

import (

)

// mp3Codec contains the codec describing MP3
const mp3Codec = "MP3"

// mp3FFmpegCodec contains the ffmpeg codec describing MP3
const mp3FFmpegCodec = "libmp3lame"

// MP3CBROptions represents the options for a MP3 CBR transcoder
type MP3CBROptions struct {
	quality string
}

// Codec returns the codec used
func (m MP3CBROptions) Codec() string {
	return mp3Codec
}

// FFmpegCodec returns the codec used by ffmpeg
func (m MP3CBROptions) FFmpegCodec() string {
	return mp3FFmpegCodec
}

// Quality returns the quality used
func (m MP3CBROptions) Quality() string {
	return m.quality
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m MP3CBROptions) FFmpegQuality() string {
	return "-something " + m.quality
}

// MP3VBROptions represents the options for a MP3 VBR transcoder
type MP3VBROptions struct {
	quality string
}

// Codec returns the codec used
func (m MP3VBROptions) Codec() string {
	return mp3Codec
}

// FFmpegCodec returns the codec used by ffmpeg
func (m MP3VBROptions) FFmpegCodec() string {
	return mp3FFmpegCodec
}

// Quality returns the quality used
func (m MP3VBROptions) Quality() string {
	return m.quality
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m MP3VBROptions) FFmpegQuality() string {
	// Return the number after 'V'
	return "-qscale:a " + string(m.quality[1])
}
