package api

import (
	"database/sql"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/transcode"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// GetTranscode returns a transcoded media file stream from wavepipe.  On success, this API will
// return a binary transcode. On failure, it will return a JSON error.
func GetTranscode(httpReq *http.Request, httpRes http.ResponseWriter, r render.Render, params martini.Params) {
	// Output struct for transcode errors
	res := ErrorResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		res.RenderError(400, "no integer transcode ID provided")
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		res.RenderError(400, "invalid integer transcode ID")
		return
	}

	// Attempt to load the song with matching ID
	song := new(data.Song)
	song.ID = id
	if err := song.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			res.RenderError(404, "song ID not found")
			return
		}

		// All other errors
		log.Println(err)
		res.ServerError()
		return
	}

	// Attempt to access data stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)
		res.ServerError()
		return
	}
	defer stream.Close()

	// Check for an input codec
	query := httpReq.URL.Query()
	codec := query.Get("codec")
	if codec == "" {
		// Default to MP3
		codec = "mp3"
	}

	// Check for an input quality
	quality := query.Get("quality")
	if quality == "" {
		// Default to 192kbps
		quality = "192"
	}

	// Create a transcoder using factory
	transcoder, err := transcode.Factory(codec, quality)
	if err != nil {
		// Check for client errors
		// Invalid codec selected
		if err == transcode.ErrInvalidCodec {
			res.RenderError(400, "invalid transcoder codec: "+codec)
			return
		} else if err == transcode.ErrInvalidQuality {
			res.RenderError(400, "invalid quality for codec "+codec+": "+quality)
			return
		}

		// All other errors, server errors
		log.Println(err)
		res.ServerError()
		return
	}

	// Set song into the transcoder
	transcoder.SetSong(song)

	// Output the command
	path := transcode.FFmpegPath
	cmd := transcoder.FFmpeg().Arguments()
	log.Println("transcode: starting:", path, cmd)

	// Invoke ffmpeg to create a transcoded audio stream
	ffmpeg := exec.Command(path, cmd...)
	mimeType := transcoder.MIMEType()

	// Generate an io.ReadCloser from ffmpeg's stdout stream
	transcode, err := ffmpeg.StdoutPipe()
	if err != nil {
		log.Println(err)
		res.ServerError()
		return
	}

	// Invoke the ffmpeg process
	if err := ffmpeg.Start(); err != nil {
		log.Println(err)
		res.ServerError()
		return
	}

	// Now that ffmpeg has started, we must assume binary data is being transferred,
	// so no more error JSON may be sent.

	// Attempt to send transcoded file stream over HTTP
	log.Printf("transcode: starting: [#%05d] %s - %s [%s %s]", song.ID, song.Artist, song.Title, transcoder.Codec(), transcoder.Quality())

	// Send transcode stream, no size for now (estimate later), set MIME type from options
	if err := httpStream(song, mimeType, -1, transcode, httpRes); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		log.Println("transcode: error:", err)
		return
	}

	log.Printf("transcode: completed: [#%05d] %s - %s", song.ID, song.Artist, song.Title)

	// Wait for ffmpeg to exit
	if err := ffmpeg.Wait(); err != nil {
		log.Println(err)
		return
	}

	return
}
