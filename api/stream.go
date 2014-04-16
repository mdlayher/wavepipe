package api

import (
	"database/sql"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// StreamResponse represents the JSON response for /api/streams
type StreamResponse struct {
	Error *Error `json:"error"`
}

// GetStream a raw, non-transcoded, media file stream from wavepipe.  On success, this API will
// return a binary stream. On failure, it will return a JSON error.
func GetStream(httpRes http.ResponseWriter, r render.Render, params martini.Params) {
	// Output struct for streams request
	res := StreamResponse{}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "unsupported API version: " + version
			r.JSON(400, res)
			return
		}
	}

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		res.Error = new(Error)
		res.Error.Code = 400
		res.Error.Message = "no integer stream ID provided"
		r.JSON(400, res)
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		res.Error = new(Error)
		res.Error.Code = 400
		res.Error.Message = "invalid integer stream ID"
		r.JSON(400, res)
		return
	}

	// Attempt to load the song with matching ID
	song := new(data.Song)
	song.ID = id
	if err := song.Load(); err != nil {
		res.Error = new(Error)

		// Check for invalid ID
		if err == sql.ErrNoRows {
			res.Error.Code = 404
			res.Error.Message = "song ID not found"
			r.JSON(404, res)
			return
		}

		// All other errors
		log.Println(err)
		res.Error.Code = 500
		res.Error.Message = "server error"
		r.JSON(500, res)
		return
	}

	// Attempt to access data stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)

		res.Error = new(Error)
		res.Error.Code = 500
		res.Error.Message = "server error"
		r.JSON(500, res)
		return
	}
	defer stream.Close()

	// Attempt to send file stream over HTTP
	log.Printf("stream: starting: [#%05d] %s - %s ", song.ID, song.Artist, song.Title)
	if err := httpStream(song, stream, httpRes); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		log.Println("stream: error:", err)
		return
	}

	log.Printf("stream: completed: [#%05d] %s - %s", song.ID, song.Artist, song.Title)
	return
}

// httpStream provides a common method to transfer a file stream using a HTTP response writer
// TODO: use this for transcoded file streams later on as well
func httpStream(song *data.Song, stream io.ReadCloser, httpRes http.ResponseWriter) error {
	// Total bytes transferred
	var total int64

	// Track the stream's progress via log
	stopProgressChan := make(chan struct{})
	go func() {
		// Track start time
		startTime := time.Now()

		// Print progress every 5 seconds
		progress := time.NewTicker(5 * time.Second)

		// Calculate total file size
		totalSize := float64(song.FileSize) / 1024 / 1024
		for {
			select {
			// Print progress
			case <-progress.C:
				// Capture current progress
				currTotal := atomic.LoadInt64(&total)
				current := float64(currTotal) / 1024 / 1024

				// Capture current percentage
				percent := int64(float64(float64(currTotal)/float64(song.FileSize)) * 100)

				// Capture current transfer rate
				rate := float64(float64((currTotal*8)/1024/1024) / float64(time.Now().Sub(startTime).Seconds()))

				log.Printf("[%d] [%03d%%] %02.3f / %02.3f MB [%02.3f Mbps]", song.ID, percent, current, totalSize, rate)
			// Stop printing
			case <-stopProgressChan:
				return
			}
		}
	}()

	// Stop progress on return
	defer func() {
		stopProgressChan <- struct{}{}
	}()

	// Buffer to store and transfer file bytes
	buf := make([]byte, 8192)

	// Indicate when stream is complete
	streamComplete := false

	// Set MIME type and content lengthfrom file, begin transferring the data stream
	httpRes.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(song.FileName)))
	httpRes.Header().Set("Content-Length", strconv.FormatInt(song.FileSize, 10))
	for {
		// Read in a buffer from the file
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			// Halt streaming after next write
			streamComplete = true
		}

		// Count bytes
		atomic.AddInt64(&total, int64(n))

		// Write bytes over HTTP
		if _, err := httpRes.Write(buf[:n]); err != nil && err != io.EOF {
			return err
		}

		// If stream is complete, break loop
		if streamComplete {
			break
		}
	}

	// No errors
	return nil
}
