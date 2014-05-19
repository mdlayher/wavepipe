wavepipe [![Build Status](https://travis-ci.org/mdlayher/wavepipe.svg?branch=master)](https://travis-ci.org/mdlayher/wavepipe) [![GoDoc](http://godoc.org/github.com/mdlayher/wavepipe?status.png)](http://godoc.org/github.com/mdlayher/wavepipe)
========

Cross-platform media server, written in Go.  MIT Licensed.

API documentation may be found in [API.md](https://github.com/mdlayher/wavepipe/blob/master/API.md).  Full code documentation may be found on [GoDoc](http://godoc.org/github.com/mdlayher/wavepipe).

wavepipe is a spiritual successor to [WaveBox](https://github.com/einsteinx2/WaveBox), and much of its design
and ideas are inspired from the WaveBox project.  This being said, wavepipe is an entirely new project, with
its own goals.

In simple terms, wavepipe will scan a media library, and expose it as a RESTful API for client consumption.
More features may be added at a later date, but first priority is to create a working API, and a web client
for consuming that API.

Installation
============

wavepipe can be built using Go 1.1+, but also has a dependency on [TagLib](https://github.com/taglib/taglib)
for its ability to read media metadata.  The TagLib static libraries can be installed on Ubuntu as follows:

`$ sudo apt-get install libtagc0-dev`

Once the TagLib library is installed, wavepipe can be downloaded, built, and installed, simply by running:

`$ go get github.com/mdlayher/wavepipe`

To enable wavepipe's transcoding functionality, you must have `ffmpeg` installed.  In order to enable MP3
and Ogg Vorbis transcoding, `ffmpeg` must have the `libmp3lame` and `libvorbis` codecs, respectively.  If
the codec is missing, transcoding to that codec will be disabled.

On newer versions of Ubuntu, `ffmpeg` with `libmp3lame` and `libvorbis` can be installed as follows:

`$ sudo apt-get install ffmpeg libavcodec-extra-53`

Configuration
=============

On first run, wavepipe will attempt to create its sqlite database and configuration file in
`~/.config/wavepipe/`.  Once this is done, the user must modify the configuration file to include a valid
media folder.  Here is an example of the `~/.config/wavepipe/wavepipe.json` configuration file, with comments.

```
{
	// The port on which wavepipe will expose its API
	"port": 8080,
	// The media folder which wavepipe will scan for valid media files
	"mediaFolder": "~/Music/",
	// Configuration for the sqlite database
	"sqlite": {
		// sqlite database file location
		"file": "~/.config/wavepipe/wavepipe.db"
	}
}
```


FAQ
===

__Q: Does wavepipe recognize the `ALBUMARTIST` tag?__

A: Not yet, but it will in the future!  I am currently developing a native Go audio tag parser, inspired by
[TagLib](https://github.com/taglib/taglib).  The project is called [taggolib](https://github.com/mdlayher/taggolib),
and I intend to use it with wavepipe to completely remove the need for Cgo and TagLib bindings.  Once taggolib is
able to parse a wide variety of media formats, the [dev_taggolib](https://github.com/mdlayher/wavepipe/tree/dev_taggolib)
branch will be merged into master, providing additional functionality and much more tagging flexibility.

__Q: Is wavepipe compatible with existing media servers?__

A: Yes, but it's a work in progress.  In order to help spur wavepipe adoption, I have started building a Subsonic
emulation layer.  This work can be found in the [dev_subsonic_api](https://github.com/mdlayher/wavepipe/tree/dev_subsonic_api)
branch, and is currently functional with the [Clementine](https://www.clementine-player.org/) media player's Subsonic
plugin.  As more of the Subsonic API is emulated, more clients will become compatible with wavepipe.
