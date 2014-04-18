package transcode

import (

)

// Options represents an audio codec and its quality settings, and includes methods to
// retrieve these settings
type Options interface {
	Codec() string
	FFmpegCodec() string
	FFmpegQuality() string
	Quality() string
}
