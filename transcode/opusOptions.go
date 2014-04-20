package transcode

// opusCodec contains the codec describing OPUS
const opusCodec = "Ogg Opus"

// opusExt contains the extension describing OPUS
const opusExt = "opus"

// opusMIMEType contains the MIME type describing OPUS
const opusMIMEType = "audio/ogg; codecs=opus"

// OPUSCBROptions represents the options for a OPUS CBR transcoder
type OPUSCBROptions struct {
	quality string
}

// Codec returns the codec used
func (m OPUSCBROptions) Codec() string {
	return opusCodec
}

// Ext returns the file extension used
func (m OPUSCBROptions) Ext() string {
	return opusExt
}

// FFmpegFlags returns the flag used by ffmpeg to signify this encoding
func (m OPUSCBROptions) FFmpegFlags() string {
	return "-ab"
}

// FFmpegCodec returns the codec used by ffmpeg
func (m OPUSCBROptions) FFmpegCodec() string {
	return FFmpegOPUSCodec
}

// MIMEType returns the MIME type of this item
func (m OPUSCBROptions) MIMEType() string {
	return opusMIMEType
}

// Quality returns the quality used
func (m OPUSCBROptions) Quality() string {
	return m.quality + "kbps"
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m OPUSCBROptions) FFmpegQuality() string {
	return m.quality + "k"
}

// OPUSVBROptions represents the options for a OPUS VBR transcoder
type OPUSVBROptions struct {
	quality string
}

// Codec returns the codec used
func (m OPUSVBROptions) Codec() string {
	return opusCodec
}

// Ext returns the file extension used
func (m OPUSVBROptions) Ext() string {
	return opusExt
}

// FFmpegCodec returns the codec used by ffmpeg
func (m OPUSVBROptions) FFmpegCodec() string {
	return FFmpegOPUSCodec
}

// FFmpegFlags returns the flag used by ffmpeg to signify this encoding
func (m OPUSVBROptions) FFmpegFlags() string {
	return "-aq"
}

// MIMEType returns the MIME type of this item
func (m OPUSVBROptions) MIMEType() string {
	return opusMIMEType
}

// Quality returns the quality used
func (m OPUSVBROptions) Quality() string {
	return m.quality
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m OPUSVBROptions) FFmpegQuality() string {
	// Return the number after 'Q'
	return string(m.quality[1:])
}
