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

// SongsResponse represents the JSON response for /api/songs
type SongsResponse struct {
	Error *Error      `json:"error"`
	Songs []data.Song `json:"songs"`
}

// GetSongs retrieves one or more songs from wavepipe, and returns a HTTP status and JSON
func GetSongs(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for songs request
	out := SongsResponse{}

	// List of songs to return
	songs := make([]data.Song, 0)

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := mux.Vars(req)["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			r.JSON(res, 400, errRes(400, "invalid integer song ID"))
			return
		}

		// Load the song
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

		// Add song to slice
		songs = append(songs, *song)
	} else {
		// Check for a limit parameter
		if pLimit := req.URL.Query().Get("limit"); pLimit != "" {
			// Split limit into two integers
			var offset int
			var count int
			if n, err := fmt.Sscanf(pLimit, "%d,%d", &offset, &count); n < 2 || err != nil {
				r.JSON(res, 400, errRes(400, "invalid comma-separated integer pair for limit"))
				return
			}

			// Retrieve limited subset of songs
			tempSongs, err := data.DB.LimitSongs(offset, count)
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy songs into the output slice
			songs = tempSongs
		} else if pRandom := req.URL.Query().Get("random"); pRandom != "" {
			// Check for a random songs request
			random, err := strconv.Atoi(pRandom)
			if err != nil {
				r.JSON(res, 400, errRes(400, "invalid integer for random"))
				return
			}

			// Retrieve the specified number of random songs
			tempSongs, err := data.DB.RandomSongs(random)
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy songs into the output slice
			songs = tempSongs
		} else {
			// Retrieve all songs
			tempSongs, err := data.DB.AllSongs()
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy songs into the output slice
			songs = tempSongs
		}
	}

	// Build response
	out.Error = nil
	out.Songs = songs

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
