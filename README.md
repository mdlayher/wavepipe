wavepipe [![Build Status](https://travis-ci.org/mdlayher/wavepipe.svg?branch=master)](https://travis-ci.org/mdlayher/wavepipe)
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

Usage
=====

At the moment, wavepipe features a very simple command-line interface.  This configuration will eventually be
replaced by a configuration file, but for the time being, command-line flags can be used for testing.

`$ wavepipe -db=wavepipe.db /media/path`

FAQ
===

__Q: Does wavepipe recognize the `ALBUMARTIST` tag?__

A: Right now, unfortunately, the answer is no.  It does not appear that this tag is supported by
[TagLib](https://github.com/taglib/taglib), the library and bindings I am currently using to extract
metadata from media.  I'd like to fix this issue in a later release, but finishing the core functionality
is my current priority.
