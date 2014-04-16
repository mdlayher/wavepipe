package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// ArtistsResponse represents the JSON response for /api/artists
type ArtistsResponse struct {
	Error   *Error        `json:"error"`
	Artists []data.Artist `json:"artists"`
	Albums  []data.Album  `json:"albums"`
	Songs   []data.Song   `json:"songs"`
}

// GetArtists retrieves one or more artists from wavepipe, and returns a HTTP status and JSON
func GetArtists(r render.Render, req *http.Request, params martini.Params) {
	// Output struct for artists request
	res := ArtistsResponse{}

	// List of artists to return
	artists := make([]data.Artist, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "unsupported API version: " + version
			r.JSON(400, res)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "invalid integer artist ID"
			r.JSON(400, res)
			return
		}

		// Load the artist
		artist := new(data.Artist)
		artist.ID = id
		if err := artist.Load(); err != nil {
			res.Error = new(Error)

			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.Error.Code = 404
				res.Error.Message = "artist ID not found"
				r.JSON(404, res)
				return
			}

			// All other errors
			log.Println(err)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// On single artist, load the albums for this artist
		albums, err := data.DB.AlbumsForArtist(artist.ID)
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Add albums to output
		res.Albums = albums

		// Get songs for this artist if parameter is true
		if req.URL.Query().Get("songs") == "true" {
			// If requested, on single artist, load the songs for this artist
			songs, err := data.DB.SongsForArtist(artist.ID)
			if err != nil {
				log.Println(err)
				res.Error = new(Error)
				res.Error.Code = 500
				res.Error.Message = "server error"
				r.JSON(500, res)
				return
			}

			// Add songs to output
			res.Songs = songs
		}

		// Add artist to slice
		artists = append(artists, *artist)
	} else {
		// Retrieve all artists
		tempArtists, err := data.DB.AllArtists()
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Copy artists into the output slice
		artists = tempArtists
	}

	// Build response
	res.Error = nil
	res.Artists = artists

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}
