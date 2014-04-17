package api

import (
	"database/sql"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// GetTranscode returns a transcoded media file stream from wavepipe.  On success, this API will
// return a binary transcode. On failure, it will return a JSON error.
func GetTranscode(httpRes http.ResponseWriter, r render.Render, params martini.Params) {
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

	// Invoke ffmpeg to create a transcoded audio stream
	ffmpeg := exec.Command("ffmpeg", "-i", song.FileName, "-codec:a", "libmp3lame", "-qscale:a", "2", "pipe:1.mp3")

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
	log.Printf("transcode: starting: [#%05d] %s - %s ", song.ID, song.Artist, song.Title)
	if err := httpStream(song, -1, transcode, httpRes); err != nil {
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
