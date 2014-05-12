package api

import (
	"bytes"
	"database/sql"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	// Extra image manipulation formats
	_ "image/jpeg"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/mdlayher/render"
	"github.com/nfnt/resize"
)

// GetArt a binary art file from wavepipe.  On success, this API will
// return binary art. On failure, it will return a JSON error.
func GetArt(httpReq *http.Request, httpRes http.ResponseWriter, r render.Render, params martini.Params) {
	// Output struct for errors
	errRes := ErrorResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			errRes.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		errRes.RenderError(400, "no integer art ID provided")
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		errRes.RenderError(400, "invalid integer art ID")
		return
	}

	// Attempt to load the art with matching ID
	art := new(data.Art)
	art.ID = id
	if err := art.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			errRes.RenderError(404, "art ID not found")
			return
		}

		// All other errors
		log.Println(err)
		errRes.ServerError()
		return
	}

	// Attempt to access art data stream
	stream, err := art.Stream()
	if err != nil {
		log.Println(err)
		errRes.ServerError()
		return
	}
	defer stream.Close()

	// Output for HTTP headers
	var length int64
	var mimeType string

	// Output art buffer
	artBuf := make([]byte, 0)

	// Check for resize request
	size := httpReq.URL.Query().Get("size")
	if size != "" {
		// Ensure size is a positive integer
		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Verify positive integer
		if sizeInt < 1 {
			errRes.RenderError(400, "negative integer size")
			return
		}

		// Decode input image stream
		img, _, err := image.Decode(stream)
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Generate a thumbnail image of the specified size
		thumb := resize.Resize(uint(sizeInt), 0, img, resize.NearestNeighbor)

		// Encode image as PNG to prevent further quality loss
		buffer := bytes.NewBuffer(make([]byte, 0))
		if err := png.Encode(buffer, thumb); err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Copy PNG into output buffer
		artBuf = buffer.Bytes()

		// Set HTTP response
		length = int64(buffer.Len())
		mimeType = "image/png"
	} else {
		// Read in the entire art stream
		tempBuf, err := ioutil.ReadAll(stream)
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Copy into output buffer
		artBuf = tempBuf

		// Set HTTP response
		length = art.FileSize
		mimeType = mime.TypeByExtension(path.Ext(art.FileName))
	}

	// Set necessary HTTP output headers

	// Get content length from file size
	httpRes.Header().Set("Content-Length", strconv.FormatInt(length, 10))

	// Set content type via MIME type
	httpRes.Header().Set("Content-Type", mimeType)

	// Get art last modify time in RFC1123 format, replace UTC with GMT
	lastMod := strings.Replace(time.Unix(art.LastModified, 0).UTC().Format(time.RFC1123), "UTC", "GMT", 1)

	// Set last modified time
	httpRes.Header().Set("Last-Modified", lastMod)

	// Specify connection close on send
	httpRes.Header().Set("Connection", "close")

	// Transfer the art stream via HTTP response writer
	if _, err := httpRes.Write(artBuf); err != nil && err != io.EOF {
		log.Println(err)
	}

	return
}
