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

// AlbumsResponse represents the JSON response for /api/albums
type AlbumsResponse struct {
	Error  *Error       `json:"error"`
	Albums []data.Album `json:"albums"`
	Songs  []data.Song  `json:"songs"`
}

// GetAlbums retrieves one or more albums from wavepipe, and returns a HTTP status and JSON
func GetAlbums(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for albums request
	out := AlbumsResponse{}

	// List of albums to return
	albums := make([]data.Album, 0)

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
			r.JSON(res, 400, errRes(400, "invalid integer album ID"))
			return
		}

		// Load the album
		album := new(data.Album)
		album.ID = id
		if err := album.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				r.JSON(res, 404, errRes(404, "album ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// On single album, load the songs for this album
		songs, err := data.DB.SongsForAlbum(album.ID)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Add songs to output
		out.Songs = songs

		// Add album to slice
		albums = append(albums, *album)
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

			// Retrieve limited subset of albums
			tempAlbums, err := data.DB.LimitAlbums(offset, count)
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy albums into the output slice
			albums = tempAlbums
		} else {
			// Retrieve all albums
			tempAlbums, err := data.DB.AllAlbums()
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy albums into the output slice
			albums = tempAlbums
		}
	}

	// Build response
	out.Error = nil
	out.Albums = albums

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
