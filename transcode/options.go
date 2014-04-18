package transcode

// Options represents an audio codec and its quality settings, and includes methods to
// retrieve these settings
type Options interface {
	Codec() string
	Ext() string
	FFmpegCodec() string
	FFmpegFlags() string
	FFmpegQuality() string
	MIMEType() string
	Quality() string
}
