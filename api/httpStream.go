package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mdlayher/wavepipe/common"
	"github.com/mdlayher/wavepipe/data"
)

var (
	// ErrCannotSeek is returned when the input stream is not seekable
	ErrCannotSeek = errors.New("httpStream: cannot seek input stream")
	// ErrInvalidRange is returned when the client attempts to retrieve an out-of-bounds part of the stream
	ErrInvalidRange = errors.New("httpStream: invalid range")
)

// HTTPStream provides a common method to transfer a file stream using a HTTP response writer
func HTTPStream(song *data.Song, mimeType string, contentLength int64, inputStream io.Reader, req *http.Request, res http.ResponseWriter) error {
	// Total bytes transferred
	var total int64

	// Output data stream, which uses the input stream by default
	stream := inputStream

	// Check for a Range header with bytes request, meaning the client is seeking through the stream
	// If client requests the entire stream (browsers, "bytes=0-"), skip range logic
	rawRange := req.Header.Get("Range")
	if rawRange != "" && rawRange != "bytes=0-" && strings.HasPrefix(rawRange, "bytes=") {
		// Check if input stream is seekable
		seekStream, ok := inputStream.(io.ReadSeeker)
		if !ok || contentLength < 0 {
			return ErrCannotSeek
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

		// Check for invalid boundaries for seeking
		if startOffset > contentLength || endOffset > contentLength {
			return ErrInvalidRange
		}

		// Seek the file stream to the starting offset
		if _, err := seekStream.Seek(startOffset, 0); err != nil {
			return err
		}

		// Recalculate content length to account for start offset
		contentLength = song.FileSize - startOffset

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

		// Track last total on each iteration, to check for zero change
		var lastTotal int64

		// Calculate total file length
		totalSize := float64(contentLength) / 1024 / 1024
		for {
			select {
			// Print progress
			case <-progress.C:
				// Capture current progress
				currTotal := atomic.LoadInt64(&total)
				current := float64(currTotal) / 1024 / 1024

				// Check if no change since last run
				if currTotal == lastTotal {
					break
				}

				// Update last total
				lastTotal = currTotal

				// Capture current transfer rate
				rate := float64(float64((currTotal*8)/1024/1024) / float64(time.Now().Sub(startTime).Seconds()))

				// If size available, we can print percentage and file sizes
				if contentLength > 0 {
					// Capture current percentage
					percent := int64(float64(float64(currTotal)/float64(contentLength)) * 100)

					log.Printf("[#%05d] [%03d%%] %02.3f / %02.3f MB [%02.3f Mbps]", song.ID, percent, current, totalSize, rate)
					break
				}

				// Else, print the current transfer size and rate
				log.Printf("[#%05d] sent: %02.3f MB [%02.3f Mbps]", song.ID, current, rate)
			// Stop printing
			case <-stopProgressChan:
				return
			}
		}
	}()

	// Stop progress on return
	defer close(stopProgressChan)

	// Set necessary output HTTP headers

	// Set Content-Length if set
	// NOTE: HTTP standards specify that this must be an exact length, so we cannot estimate it for
	// transcodes unless the entire file is transcoded and then sent
	if contentLength > 0 {
		res.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}

	// Override Content-Type if set
	contentType := mime.TypeByExtension(path.Ext(song.FileName))
	if mimeType != "" {
		contentType = mimeType
	}
	res.Header().Set("Content-Type", contentType)

	// Set Last-Modified using filesystem modify time
	res.Header().Set("Last-Modified", common.UNIXtoRFC1123(song.LastModified))

	// Specify connection close on send
	res.Header().Set("Connection", "close")

	// Begin transferring the data stream
	for {
		// Copy bytes in chunks from input stream to output HTTP response
		n, err := io.CopyN(res, stream, 8192)
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			return nil
		}

		// Count bytes sent to track progress
		atomic.AddInt64(&total, int64(n))
	}
}
