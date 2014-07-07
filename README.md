wavepipe [![Build Status](https://travis-ci.org/mdlayher/wavepipe.svg?branch=master)](https://travis-ci.org/mdlayher/wavepipe) [![GoDoc](http://godoc.org/github.com/mdlayher/wavepipe?status.svg)](http://godoc.org/github.com/mdlayher/wavepipe)
========

Cross-platform media server, written in Go.  MIT Licensed.

API documentation may be found in [API.md](https://github.com/mdlayher/wavepipe/blob/master/API.md).
Full code documentation may be found on [GoDoc](http://godoc.org/github.com/mdlayher/wavepipe).

wavepipe is a spiritual successor to [WaveBox](https://github.com/einsteinx2/WaveBox), and much of its design
and ideas are inspired from the WaveBox project.  This being said, wavepipe is an entirely new project, with
its own goals.

In simple terms, wavepipe will scan a media library, and expose it as a RESTful API for client consumption.
More features may be added at a later date, but first priority is to create a working API, and a web client
for consuming that API.

Installation
============

wavepipe can be built using Go 1.3+, but also has a dependency on [TagLib](https://github.com/taglib/taglib)
for its ability to read media metadata.  The TagLib static libraries can be installed on Ubuntu as follows:

`$ sudo apt-get install libtagc0-dev`

Once the TagLib library is installed, wavepipe can be downloaded, built, and installed, simply by running:

`$ go get github.com/mdlayher/wavepipe`

To aid in debugging, the current git commit revision can be injected into wavepipe via the Go linker. If wavepipe
is built without the proper flags, it will log a warning stating "empty git revision", and ask to be built using
`make`.  For this reason, it is recommended to build and install wavepipe using the included `Makefile`:

```
$ make
$ make install
```

To enable wavepipe's transcoding functionality, you must have `ffmpeg` installed.  In order to enable MP3
and Ogg Vorbis transcoding, `ffmpeg` must have the `libmp3lame` and `libvorbis` codecs, respectively.  If
the codec is missing, transcoding to that codec will be disabled.

On newer versions of Ubuntu, `ffmpeg` with `libmp3lame` and `libvorbis` can be installed as follows:

`$ sudo apt-get install ffmpeg libavcodec-extra-53`

Configuration
=============

On first run, wavepipe will attempt to create its sqlite database using the option set via the `-sqlite` flag.
The default location is `~/.config/wavepipe/wavepipe.db`.  Once this is done, the user must at least specify
the `-media` command line flag, to allow wavepipe to scan and watch a media folder.  Here is an example of
the default command-line configuration, with the media folder specified as the user's home media folder:

```
$ wavepipe -host :8080 -sqlite ~/.config/wavepipe/wavepipe.db -media ~/Music/
```

These options will:
  - Bind wavepipe to localhost on port 8080
  - Use the specified location for its sqlite database
  - Scan and watch media in the specified folder

The host and sqlite database will use the above configuration by default, but the media folder must be specified.

```
$ wavepipe -media ~/Music/
```

Recommendations
===============

For the best possible experience while using wavepipe, it is recommended that you follow these tips:
  - Run wavepipe using SSL.  In order to effectively secure your wavepipe session, SSL is a must.
    The most simple way to accomplish this is to proxy requests to wavepipe via [nginx](http://nginx.org), with nginx
    configured to use SSL.
  - Ensure your media is properly tagged.  wavepipe will provide a much better experience for users who ensure that their
    media is consistently tagged.  Proper artist and album naming are especially key, in order to enable the best possible
    experience.

FAQ
===

__Q: Where's the web UI?__

A: My frontend skills are very limited, but I am in the process of learning [Ember.js](http://emberjs.com).  If you'd
like to help contribute code for an official wavepipe web UI, I'd love to hear from you!  Feel free to contact me
at mdlayher (at) gmail (dot) com!

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
