package transcode

// mp3Codec contains the codec describing MP3
const mp3Codec = "MP3"

// mp3Ext contains the extension describing MP3
const mp3Ext = "mp3"

// mp3FFmpegCodec contains the ffmpeg codec describing MP3
const mp3FFmpegCodec = "libmp3lame"

// mp3MIMEType contains the MIME type describing MP3
const mp3MIMEType = "audio/mpeg"

// MP3CBROptions represents the options for a MP3 CBR transcoder
type MP3CBROptions struct {
	quality string
}

// Codec returns the codec used
func (m MP3CBROptions) Codec() string {
	return mp3Codec
}

// Ext returns the file extension used
func (m MP3CBROptions) Ext() string {
	return mp3Ext
}

// FFmpegFlags returns the flag used by ffmpeg to signify this encoding
func (m MP3CBROptions) FFmpegFlags() string {
	return "-b:a"
}

// FFmpegCodec returns the codec used by ffmpeg
func (m MP3CBROptions) FFmpegCodec() string {
	return mp3FFmpegCodec
}

// MIMEType returns the MIME type of this item
func (m MP3CBROptions) MIMEType() string {
	return mp3MIMEType
}

// Quality returns the quality used
func (m MP3CBROptions) Quality() string {
	return m.quality + "kbps"
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m MP3CBROptions) FFmpegQuality() string {
	return m.quality + "k"
}

// MP3VBROptions represents the options for a MP3 VBR transcoder
type MP3VBROptions struct {
	quality string
}

// Codec returns the codec used
func (m MP3VBROptions) Codec() string {
	return mp3Codec
}

// Ext returns the file extension used
func (m MP3VBROptions) Ext() string {
	return mp3Ext
}

// FFmpegCodec returns the codec used by ffmpeg
func (m MP3VBROptions) FFmpegCodec() string {
	return mp3FFmpegCodec
}

// FFmpegFlags returns the flag used by ffmpeg to signify this encoding
func (m MP3VBROptions) FFmpegFlags() string {
	return "-qscale:a"
}

// MIMEType returns the MIME type of this item
func (m MP3VBROptions) MIMEType() string {
	return mp3MIMEType
}

// Quality returns the quality used
func (m MP3VBROptions) Quality() string {
	return m.quality
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m MP3VBROptions) FFmpegQuality() string {
	// Return the number after 'V'
	return string(m.quality[1:])
}
