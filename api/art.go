package api

import (
	"bytes"
	"database/sql"
	"image"
	"image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"

	// Extra image manipulation formats
	_ "image/jpeg"

	"github.com/mdlayher/wavepipe/common"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/unrolled/render"
)

// GetArt retrieves a binary art file from wavepipe, optionally resizing the art file.
// On success, this API will return binary art. On failure, it will return a JSON error.
func GetArt(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(r)["id"]
	if !ok {
		ren.JSON(w, 400, errRes(400, "no integer art ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		ren.JSON(w, 400, errRes(400, "invalid integer art ID"))
		return
	}

	// Attempt to load the art with matching ID
	art := &data.Art{ID: id}
	if err := art.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			ren.JSON(w, 404, errRes(404, "art ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Attempt to access art data stream
	stream, err := art.Stream()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}
	defer stream.Close()

	// Output stream
	var outStream io.Reader

	// Output for HTTP headers
	var length int64
	var mimeType string

	// Check for resize request
	if size := r.URL.Query().Get("size"); size != "" {
		// Ensure size is a valid integer
		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			ren.JSON(w, 400, errRes(400, "invalid integer size"))
			return
		}

		// Verify positive integer
		if sizeInt < 1 {
			ren.JSON(w, 400, errRes(400, "negative integer size"))
			return
		}

		// Decode input image stream
		img, _, err := image.Decode(stream)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Generate a thumbnail image of the specified size
		thumb := resize.Resize(uint(sizeInt), 0, img, resize.NearestNeighbor)

		// Encode image as PNG to prevent further quality loss
		buffer := bytes.NewBuffer(make([]byte, 0))
		if err := png.Encode(buffer, thumb); err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Set HTTP response
		length = int64(buffer.Len())
		mimeType = "image/png"

		// Store the resized art stream for output
		outStream = buffer
	} else {
		// If not resizing, set HTTP headers with known values
		length = art.FileSize
		mimeType = mime.TypeByExtension(path.Ext(art.FileName))

		// Store the original art stream for output
		outStream = stream
	}

	// Set necessary HTTP output headers

	// Get content length from file size
	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))

	// Set content type via MIME type
	w.Header().Set("Content-Type", mimeType)

	// Set last modified time in RFC1123 format
	w.Header().Set("Last-Modified", common.UNIXtoRFC1123(art.LastModified))

	// Specify connection close on send
	w.Header().Set("Connection", "close")

	// Stream the output over HTTP
	if _, err := io.Copy(w, outStream); err != nil {
		log.Println(err)
		return
	}
}
