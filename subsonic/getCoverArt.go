package subsonic

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

	// Extra image manipulation formats
	_ "image/jpeg"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/common"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/nfnt/resize"
	"github.com/unrolled/render"
)

// GetCoverArt is used in Subsonic to retrieve cover art, specifying an ID
// and a size
func GetCoverArt(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Fetch ID parameter
	pID := req.URL.Query().Get("id")
	if pID == "" {
		r.XML(res, 200, ErrMissingParameter)
		return
	}

	// Parse ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		r.XML(res, 200, ErrMissingParameter)
		return
	}

	// Load art by ID
	art := &data.Art{ID: id}
	if err := art.Load(); err != nil {
		// If no art found, return 404
		if err == sql.ErrNoRows {
			r.XML(res, 404, nil)
			return
		}

		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// TODO: remove code duplication from api/art where possible

	// Attempt to access art data stream
	stream, err := art.Stream()
	if err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}
	defer stream.Close()

	// Output for HTTP headers
	var length int64
	var mimeType string

	// Output art buffer
	artBuf := make([]byte, 0)

	// Check for resize request
	if size := req.URL.Query().Get("size"); size != "" {
		// Ensure size is a valid integer
		sizeInt, err := strconv.Atoi(size)
		if err != nil {
			r.XML(res, 200, ErrMissingParameter)
			return
		}

		// Verify positive integer
		if sizeInt < 1 {
			r.XML(res, 200, ErrMissingParameter)
			return
		}

		// Decode input image stream
		img, _, err := image.Decode(stream)
		if err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
			return
		}

		// Generate a thumbnail image of the specified size
		thumb := resize.Resize(uint(sizeInt), 0, img, resize.NearestNeighbor)

		// Encode image as PNG to prevent further quality loss
		buffer := bytes.NewBuffer(make([]byte, 0))
		if err := png.Encode(buffer, thumb); err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
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
			r.XML(res, 200, ErrGeneric)
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

	// Set last modified time
	res.Header().Set("Last-Modified", common.UNIXtoRFC1123(art.LastModified))

	// Specify connection close on send
	res.Header().Set("Connection", "close")

	// Transfer the art stream via HTTP response writer
	if _, err := res.Write(artBuf); err != nil && err != io.EOF {
		log.Println(err)
	}

	return
}
