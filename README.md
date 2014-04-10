wavepipe [![Build Status](https://travis-ci.org/mdlayher/wavepipe.svg?branch=master)](https://travis-ci.org/mdlayher/wavepipe)
========

Cross-platform media server, written in Go.  MIT Licensed.

Full documentation may be found on [GoDoc](http://godoc.org/github.com/mdlayher/wavepipe).

wavepipe is a spiritual successor to [WaveBox](https://github.com/einsteinx2/WaveBox), and much of its design
and ideas are inspired from the WaveBox project.  This being said, wavepipe is an entirely new project, with
its own goals.

In simple terms, wavepipe will scan a media library, and expose it as a RESTful API for client consumption.
More features may be added at a later date, but first priority is to create a working API, and a web client
for consuming that API.

Installation
============

wavepipe can be built using Go 1.1+. It can be downloaded, built, and installed, simply by running:

`$ go get github.com/mdlayher/wavepipe`
