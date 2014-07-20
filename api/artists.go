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

// ArtistsResponse represents the JSON response for the Artists API.
type ArtistsResponse struct {
	Error   *Error        `json:"error"`
	Artists []data.Artist `json:"artists"`
	Albums  []data.Album  `json:"albums"`
	Songs   []data.Song   `json:"songs"`
}

// GetArtists retrieves one or more artists from wavepipe, and returns a HTTP status and JSON.
// It can be used to fetch a single artist, a limited subset of artists, or all artists, depending
// on the request parameters.
func GetArtists(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Output struct for artists request
	out := ArtistsResponse{}

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
			ren.JSON(w, 400, errRes(400, "invalid integer artist ID"))
			return
		}

		// Load the artist
		artist := &data.Artist{ID: id}
		if err := artist.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				ren.JSON(w, 404, errRes(404, "artist ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Store artist in slice
		out.Artists = []data.Artist{*artist}

		// On single artist, load the albums for this artist
		albums, err := data.DB.AlbumsForArtist(artist.ID)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Add albums to output
		out.Albums = albums

		// Get songs for this artist if parameter is true
		if r.URL.Query().Get("songs") == "true" {
			// If requested, on single artist, load the songs for this artist
			songs, err := data.DB.SongsForArtist(artist.ID)
			if err != nil {
				log.Println(err)
				ren.JSON(w, 500, serverErr)
				return
			}

			// Add songs to output
			out.Songs = songs
		}

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

		// Retrieve limited subset of artists
		artists, err := data.DB.LimitArtists(offset, count)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Artists = artists
		ren.JSON(w, 200, out)
		return
	}

	// If no other case, retrieve all artists
	artists, err := data.DB.AllArtists()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Artists = artists
	ren.JSON(w, 200, out)
	return
}
