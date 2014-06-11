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

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/unrolled/render"
)

// GetArt a binary art file from wavepipe.  On success, this API will
// return binary art. On failure, it will return a JSON error.
func GetArt(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(req)["id"]
	if !ok {
		r.JSON(res, 400, errRes(400, "no integer art ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		r.JSON(res, 400, errRes(400, "invalid integer art ID"))
		return
	}

	// Attempt to load the art with matching ID
	art := new(data.Art)
	art.ID = id
	if err := art.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			r.JSON(res, 400, errRes(400, "art ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Attempt to access art data stream
	stream, err := art.Stream()
	if err != nil {
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}
	defer stream.Close()

	// Output for HTTP headers
	var length int64
	var mimeType string

	// Output art buffer
	artBuf := make([]byte, 0)

	// Check for resize request
	size := req.URL.Query().Get("size")
	if size != "" {
		// Ensure size is a positive integer
		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			log.Println(err)
			return
		}

		// Verify positive integer
		if sizeInt < 1 {
			r.JSON(res, 400, errRes(400, "negative integer size"))
			return
		}

		// Decode input image stream
		img, _, err := image.Decode(stream)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Generate a thumbnail image of the specified size
		thumb := resize.Resize(uint(sizeInt), 0, img, resize.NearestNeighbor)

		// Encode image as PNG to prevent further quality loss
		buffer := bytes.NewBuffer(make([]byte, 0))
		if err := png.Encode(buffer, thumb); err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
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
			r.JSON(res, 500, serverErr)
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
	res.Header().Set("Content-Length", strconv.FormatInt(length, 10))

	// Set content type via MIME type
	res.Header().Set("Content-Type", mimeType)

	// Get art last modify time in RFC1123 format, replace UTC with GMT
	lastMod := strings.Replace(time.Unix(art.LastModified, 0).UTC().Format(time.RFC1123), "UTC", "GMT", 1)

	// Set last modified time
	res.Header().Set("Last-Modified", lastMod)

	// Specify connection close on send
	res.Header().Set("Connection", "close")

	// Transfer the art stream via HTTP response writer
	if _, err := res.Write(artBuf); err != nil && err != io.EOF {
		log.Println(err)
	}

	return
}
