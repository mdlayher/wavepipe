package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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

	// Attempt to serve art, handling resizing from HTTP request parameters
	if err := ServeArt(w, r, art); err != nil {
		// Client-facing errors
		if err == ErrInvalidIntegerSize || err == ErrNegativeIntegerSize {
			ren.JSON(w, 400, errRes(400, err.Error()))
			return
		}

		// Server-side errors
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}
}
