package api

import (
	"bytes"
	"database/sql"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/waveform"
	"github.com/unrolled/render"
)

// waveformCache stores encoded waveform images in-memory, for re-use
// through multiple HTTP calls
var waveformCache = map[int][]byte{}

// GetWaveform generates and returns a waveform image from wavepipe.  On success, this API will
// return a binary stream. On failure, it will return a JSON error.
func GetWaveform(w http.ResponseWriter, r *http.Request) {
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
		ren.JSON(w, 400, errRes(400, "no integer song ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		ren.JSON(w, 400, errRes(400, "invalid integer song ID"))
		return
	}

	// Attempt to load the song with matching ID
	song := &data.Song{ID: id}
	if err := song.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			ren.JSON(w, 404, errRes(404, "song ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Check for a cached waveform
	if _, ok := waveformCache[id]; ok {
		// Send cached data to HTTP writer
		if _, err := io.Copy(w, bytes.NewReader(waveformCache[id])); err != nil {
			log.Println(err)
		}

		return
	}

	// Open song's backing stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Generate a waveform from this song
	img, err := waveform.New(stream, &waveform.Options{
		ScaleX:     2,
		ScaleY:     2,
		Resolution: 2,
		ScaleRMS:   true,
		Sharpness:  1,
	})
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Encode as PNG into buffer
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		log.Println(err)
	}

	// Store cached image
	waveformCache[id] = buf.Bytes()

	// Send over HTTP
	if _, err := io.Copy(w, buf); err != nil {
		log.Println(err)
	}
}
