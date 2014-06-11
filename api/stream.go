package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// GetStream a raw, non-transcoded, media file stream from wavepipe.  On success, this API will
// return a binary stream. On failure, it will return a JSON error.
func GetStream(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Advertise that clients may send Range requests
	res.Header().Set("Accept-Ranges", "bytes")

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
		r.JSON(res, 400, errRes(400, "no integer stream ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		r.JSON(res, 400, errRes(400, "invalid integer stream ID"))
		return
	}

	// Attempt to load the song with matching ID
	song := new(data.Song)
	song.ID = id
	if err := song.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			r.JSON(res, 404, errRes(404, "song ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Attempt to access data stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Generate a string used for logging this operation
	opStr := fmt.Sprintf("[#%05d] %s - %s [%s %dkbps]", song.ID, song.Artist, song.Title,
		data.CodecMap[song.FileTypeID], song.Bitrate)

	// Attempt to send file stream over HTTP
	log.Println("stream: starting:", opStr)

	// Pass stream using song's file size, auto-detect MIME type
	if err := HTTPStream(song, "", song.FileSize, stream, req, res); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		// Check for invalid range, return HTTP 416
		if err == ErrInvalidRange {
			r.JSON(res, 416, errRes(416, "invalid HTTP Range header boundaries"))
			return
		}

		log.Println("stream: error:", err)
		return
	}

	log.Println("stream: completed:", opStr)
	return
}
