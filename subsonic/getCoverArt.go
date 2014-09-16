package subsonic

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
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

	// Attempt to serve art, handling resizing from HTTP request parameters
	if err := api.ServeArt(res, req, art); err != nil {
		// Client-facing errors
		if err == api.ErrInvalidIntegerSize || err == api.ErrNegativeIntegerSize {
			r.XML(res, 200, ErrMissingParameter)
			return
		}

		// Server-side errors
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}
}
