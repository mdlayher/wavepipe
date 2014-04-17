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
	render  render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (a *ArtistsResponse) RenderError(code int, message string) {
	// Nullify all other fields
	a.Artists = nil
	a.Albums = nil
	a.Songs = nil

	// Generate error
	a.Error = new(Error)
	a.Error.Code = code
	a.Error.Message = message

	// Render with specified HTTP status code
	a.render.JSON(code, a)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (a *ArtistsResponse) ServerError() {
	a.RenderError(500, "server error")
	return
}

// GetArtists retrieves one or more artists from wavepipe, and returns a HTTP status and JSON
func GetArtists(r render.Render, req *http.Request, params martini.Params) {
	// Output struct for artists request
	res := ArtistsResponse{render: r}

	// List of artists to return
	artists := make([]data.Artist, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.RenderError(400, "invalid integer artist ID")
			return
		}

		// Load the artist
		artist := new(data.Artist)
		artist.ID = id
		if err := artist.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.RenderError(404, "artist ID not found")
				return
			}

			// All other errors
			log.Println(err)
			res.ServerError()
			return
		}

		// On single artist, load the albums for this artist
		albums, err := data.DB.AlbumsForArtist(artist.ID)
		if err != nil {
			log.Println(err)
			res.ServerError()
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
				res.ServerError()
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
			res.ServerError()
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
