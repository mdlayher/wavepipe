package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// AlbumsResponse represents the JSON response for the Albums API.
type AlbumsResponse struct {
	Error  *Error       `json:"error"`
	Albums []data.Album `json:"albums"`
	Songs  []data.Song  `json:"songs"`
}

// GetAlbums retrieves one or more albums from wavepipe, and returns a HTTP status and JSON.
// It can be used to fetch a single album, a limited subset of albums, or all albums, depending
// on the request parameters.
func GetAlbums(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Output struct for albums request
	out := AlbumsResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := mux.Vars(r)["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			ren.JSON(w, 400, errRes(400, "invalid integer album ID"))
			return
		}

		// Load the album
		album := &data.Album{ID: id}
		if err := album.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				ren.JSON(w, 404, errRes(404, "album ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// On single album, load the songs for this album
		songs, err := data.DB.SongsForAlbum(album.ID)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Add album and songs to output
		out.Albums = []data.Album{*album}
		out.Songs = songs

		// HTTP 200 OK with JSON
		ren.JSON(w, 200, out)
		return
	}

	// Check for a limit parameter
	if pLimit := r.URL.Query().Get("limit"); pLimit != "" {
		// Split limit into two integers
		var offset int
		var count int
		if n, err := fmt.Sscanf(pLimit, "%d,%d", &offset, &count); n < 2 || err != nil {
			ren.JSON(w, 400, errRes(400, "invalid comma-separated integer pair for limit"))
			return
		}

		// Retrieve limited subset of albums
		albums, err := data.DB.LimitAlbums(offset, count)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Albums = albums
		ren.JSON(w, 200, out)
		return
	}

	// If no other case, retrieve all albums
	albums, err := data.DB.AllAlbums()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Albums = albums
	ren.JSON(w, 200, out)
	return
}
