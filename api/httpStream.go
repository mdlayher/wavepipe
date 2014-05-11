package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mdlayher/wavepipe/data"
)

// HTTPStream provides a common method to transfer a file stream using a HTTP response writer
func HTTPStream(song *data.Song, contentLength int64, inputStream io.Reader, req *http.Request, res http.ResponseWriter) error {
	// Total bytes transferred
	var total int64

	// Output data stream
	var stream io.Reader = inputStream

	// Check for a Range header with bytes request, meaning the client is seeking through the stream
	rawRange := req.Header.Get("Range")
	if rawRange != "" && strings.HasPrefix(rawRange, "bytes=") {
		// Check if input stream is seekable
		seekStream, ok := inputStream.(io.ReadSeeker)
		if !ok {
			return errors.New("cannot seek")
		}

		// Attempt to parse byte range
		pair := strings.Split(rawRange[6:], "-")

		// Parse first element as the starting point
		startOffset, err := strconv.ParseInt(pair[0], 10, 64)
		if err != nil {
			return err
		}

		// Parse second element as the ending point, if available
		var endOffset int64
		if pair[1] != "" {
			tempEndOffset, err := strconv.ParseInt(pair[1], 10, 64)
			if err != nil {
				return err
			}

			endOffset = tempEndOffset
		}

		// Seek the file stream to the starting offset
		if _, err := seekStream.Seek(startOffset, 0); err != nil {
			return err
		}

		// By default, use the length of the file (minus 1 byte) as the ending offset
		rangeEnd := contentLength - 1

		// If an ending offset was set, recalculate values, and limit the stream to return up to that point
		if endOffset > 0 {
			// Recalculate range ending offset using the specified end point
			rangeEnd = endOffset

			// Recalculate content length
			contentLength = endOffset - startOffset

			// Wrap the stream to return only contentLength bytes
			stream = io.LimitReader(seekStream.(io.Reader), contentLength)
		}

		// Respond with HTTP 206 Partial Content, and Content-Range header indicating the stream offset
		res.WriteHeader(206)
		res.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startOffset, rangeEnd, contentLength))
	}

	// Track the stream's progress via log
	stopProgressChan := make(chan struct{})
	go func() {
		// Track start time
		startTime := time.Now()

		// Print progress every 5 seconds
		progress := time.NewTicker(5 * time.Second)

		// Calculate total file length
		totalSize := float64(contentLength) / 1024 / 1024
		for {
			select {
			// Print progress
			case <-progress.C:
				// Capture current progress
				currTotal := atomic.LoadInt64(&total)
				current := float64(currTotal) / 1024 / 1024

				// Capture current transfer rate
				rate := float64(float64((currTotal*8)/1024/1024) / float64(time.Now().Sub(startTime).Seconds()))

				// If size available, we can print percentage and file sizes
				if contentLength > 0 {
					// Capture current percentage
					percent := int64(float64(float64(currTotal)/float64(contentLength)) * 100)

					log.Printf("[%d] [%03d%%] %02.3f / %02.3f MB [%02.3f Mbps]", song.ID, percent, current, totalSize, rate)
					break
				}

				// Else, print the current transfer size and rate
				log.Printf("[%d] sent: %02.3f MB [%02.3f Mbps]", song.ID, current, rate)
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
	// NOTE: HTTP standards specify that this must be an exact length, so we cannot estimate it for
	// transcodes unless the entire file is transcoded and then sent
	if contentLength > 0 {
		res.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}

	// Get song modify time in RFC1123 format, replace UTC with GMT
	lastMod := strings.Replace(time.Unix(song.LastModified, 0).UTC().Format(time.RFC1123), "UTC", "GMT", 1)

	// Set Last-Modified using filesystem modify time
	res.Header().Set("Last-Modified", lastMod)

	// Specify connection close on send
	res.Header().Set("Connection", "close")

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
		if _, err := res.Write(buf[:n]); err != nil && err != io.EOF {
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
