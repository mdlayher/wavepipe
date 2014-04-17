package api

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/mdlayher/wavepipe/data"
)

// httpStream provides a common method to transfer a file stream using a HTTP response writer
// TODO: use this for transcoded file streams later on as well
func httpStream(song *data.Song, fileSize int64, stream io.ReadCloser, httpRes http.ResponseWriter) error {
	// Total bytes transferred
	var total int64

	// Track the stream's progress via log
	stopProgressChan := make(chan struct{})
	go func() {
		// If no file size set, no point in printing progress
		if fileSize < 0 {
			return
		}

		// Track start time
		startTime := time.Now()

		// Print progress every 5 seconds
		progress := time.NewTicker(5 * time.Second)

		// Calculate total file size
		totalSize := float64(fileSize) / 1024 / 1024
		for {
			select {
			// Print progress
			case <-progress.C:
				// Capture current progress
				currTotal := atomic.LoadInt64(&total)
				current := float64(currTotal) / 1024 / 1024

				// Capture current percentage
				percent := int64(float64(float64(currTotal)/float64(fileSize)) * 100)

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
		close(stopProgressChan)
	}()

	// Buffer to store and transfer file bytes
	buf := make([]byte, 8192)

	// Indicate when stream is complete
	streamComplete := false

	// Set necessary output HTTP headers

	// Set Content-Length if set
	if fileSize > 0 {
		httpRes.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	}

	// Set Last-Modified using filesystem modify time
	httpRes.Header().Set("Last-Modified", time.Unix(song.LastModified, 0).UTC().Format(time.RFC1123))

	// Begin transferring the data stream
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
