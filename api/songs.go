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

// SongsResponse represents the JSON response for the Songs API.
type SongsResponse struct {
	Error *Error      `json:"error"`
	Songs []data.Song `json:"songs"`
}

// GetSongs retrieves one or more songs from wavepipe, and returns a HTTP status and JSON.
// It can be used to fetch a single song, a limited subset of songs, a specified number of
// random songs, or all songs, depending on the request parameters.
func GetSongs(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Output struct for songs request
	out := SongsResponse{}

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
			ren.JSON(w, 400, errRes(400, "invalid integer song ID"))
			return
		}

		// Load the song
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

		// Add song to output
		out.Songs = []data.Song{*song}

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

		// Retrieve limited subset of songs
		songs, err := data.DB.LimitSongs(offset, count)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Songs = songs
		ren.JSON(w, 200, out)
		return
	}

	// Check for a random songs request
	if pRandom := r.URL.Query().Get("random"); pRandom != "" {
		// Verify valid integer random count
		random, err := strconv.Atoi(pRandom)
		if err != nil {
			ren.JSON(w, 400, errRes(400, "invalid integer for random"))
			return
		}

		// Retrieve the specified number of random songs
		songs, err := data.DB.RandomSongs(random)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Songs = songs
		ren.JSON(w, 200, out)
		return
	}

	// If no other case, retrieve all songs
	songs, err := data.DB.AllSongs()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Songs = songs
	ren.JSON(w, 200, out)
	return
}
