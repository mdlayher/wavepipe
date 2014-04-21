package api

import (
	"database/sql"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// GetArt a binary art file from wavepipe.  On success, this API will
// return binary art. On failure, it will return a JSON error.
func GetArt(httpRes http.ResponseWriter, r render.Render, params martini.Params) {
	// Output struct for art errors
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
		res.RenderError(400, "no integer art ID provided")
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		res.RenderError(400, "invalid integer art ID")
		return
	}

	// Attempt to load the art with matching ID
	art := new(data.Art)
	art.ID = id
	if err := art.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			res.RenderError(404, "art ID not found")
			return
		}

		// All other errors
		log.Println(err)
		res.ServerError()
		return
	}

	// Attempt to access art data stream
	stream, err := art.Stream()
	if err != nil {
		log.Println(err)
		res.ServerError()
		return
	}
	defer stream.Close()

	// Read in the entire art stream
	artBuf, err := ioutil.ReadAll(stream)
	if err != nil {
		log.Println(err)
		res.ServerError()
		return
	}

	// Set necessary HTTP output headers

	// Get content length from file size
	httpRes.Header().Set("Content-Length", strconv.FormatInt(art.FileSize, 10))

	// Set content type via MIME type
	httpRes.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(art.FileName)))

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
