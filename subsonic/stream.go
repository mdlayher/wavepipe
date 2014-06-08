package subsonic

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/martini-contrib/render"
)

// GetStream is used to return the media stream for a single file
func GetStream(req *http.Request, res http.ResponseWriter, r render.Render) {
	// Fetch ID parameter
	pID := req.URL.Query().Get("id")
	if pID == "" {
		r.XML(200, ErrMissingParameter)
		return
	}

	// Parse ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Load song by ID
	song := &data.Song{ID: id}
	if err := song.Load(); err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Open file stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Generate a string used for logging this operation
	opStr := fmt.Sprintf("[#%05d] %s - %s [%s %dkbps]", song.ID, song.Artist, song.Title,
		data.CodecMap[song.FileTypeID], song.Bitrate)

	// Attempt to send file stream over HTTP
	log.Println("stream: starting:", opStr)

	// Pass stream using song's file size, auto-detect MIME type
	if err := api.HTTPStream(song, "", song.FileSize, stream, req, res); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		log.Println("stream: error:", err)
		return
	}

	log.Println("stream: completed:", opStr)
	return
}
