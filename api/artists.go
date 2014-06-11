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

// ArtistsResponse represents the JSON response for /api/artists
type ArtistsResponse struct {
	Error   *Error        `json:"error"`
	Artists []data.Artist `json:"artists"`
	Albums  []data.Album  `json:"albums"`
	Songs   []data.Song   `json:"songs"`
}

// GetArtists retrieves one or more artists from wavepipe, and returns a HTTP status and JSON
func GetArtists(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for artists request
	out := ArtistsResponse{}

	// List of artists to return
	artists := make([]data.Artist, 0)

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
			r.JSON(res, 400, errRes(400, "invalid integer artist ID"))
			return
		}

		// Load the artist
		artist := new(data.Artist)
		artist.ID = id
		if err := artist.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				r.JSON(res, 400, errRes(400, "artist ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// On single artist, load the albums for this artist
		albums, err := data.DB.AlbumsForArtist(artist.ID)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Add albums to output
		out.Albums = albums

		// Get songs for this artist if parameter is true
		if req.URL.Query().Get("songs") == "true" {
			// If requested, on single artist, load the songs for this artist
			songs, err := data.DB.SongsForArtist(artist.ID)
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Add songs to output
			out.Songs = songs
		}

		// Add artist to slice
		artists = append(artists, *artist)
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

			// Retrieve limited subset of artists
			tempArtists, err := data.DB.LimitArtists(offset, count)
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy artists into the output slice
			artists = tempArtists
		} else {
			// Retrieve all artists
			tempArtists, err := data.DB.AllArtists()
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy artists into the output slice
			artists = tempArtists
		}
	}

	// Build response
	out.Error = nil
	out.Artists = artists

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
