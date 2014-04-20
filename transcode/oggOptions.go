package transcode

// oggCodec contains the codec describing OGG
const oggCodec = "Ogg Vorbis"

// oggExt contains the extension describing OGG
const oggExt = "ogg"

// oggMIMEType contains the MIME type describing OGG
const oggMIMEType = "audio/ogg"

// OGGCBROptions represents the options for a OGG CBR transcoder
type OGGCBROptions struct {
	quality string
}

// Codec returns the codec used
func (m OGGCBROptions) Codec() string {
	return oggCodec
}

// Ext returns the file extension used
func (m OGGCBROptions) Ext() string {
	return oggExt
}

// FFmpegFlags returns the flag used by ffmpeg to signify this encoding
func (m OGGCBROptions) FFmpegFlags() string {
	return "-ab"
}

// FFmpegCodec returns the codec used by ffmpeg
func (m OGGCBROptions) FFmpegCodec() string {
	return FFMpegOGGCodec
}

// MIMEType returns the MIME type of this item
func (m OGGCBROptions) MIMEType() string {
	return oggMIMEType
}

// Quality returns the quality used
func (m OGGCBROptions) Quality() string {
	return m.quality + "kbps"
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m OGGCBROptions) FFmpegQuality() string {
	return m.quality + "k"
}

// OGGVBROptions represents the options for a OGG VBR transcoder
type OGGVBROptions struct {
	quality string
}

// Codec returns the codec used
func (m OGGVBROptions) Codec() string {
	return oggCodec
}

// Ext returns the file extension used
func (m OGGVBROptions) Ext() string {
	return oggExt
}

// FFmpegCodec returns the codec used by ffmpeg
func (m OGGVBROptions) FFmpegCodec() string {
	return FFMpegOGGCodec
}

// FFmpegFlags returns the flag used by ffmpeg to signify this encoding
func (m OGGVBROptions) FFmpegFlags() string {
	return "-aq"
}

// MIMEType returns the MIME type of this item
func (m OGGVBROptions) MIMEType() string {
	return oggMIMEType
}

// Quality returns the quality used
func (m OGGVBROptions) Quality() string {
	return m.quality
}

// FFmpegQuality returns the quality flag used by ffmpeg
func (m OGGVBROptions) FFmpegQuality() string {
	// Return the number after 'Q'
	return string(m.quality[1:])
}
