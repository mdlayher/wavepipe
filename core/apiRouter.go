package core

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
)

// apiRouter sets up the instance of martini
func apiRouter(apiKillChan chan struct{}) {
	log.Println("api: starting...")

	// Initialize martini
	m := martini.New()

	// Set up middleware
	// GZIP all requests to drastically reduce size
	m.Use(gzip.All())
	m.Use(render.Renderer())

	// Enable graceful shutdown when triggered by manager
	stopAPI := false
	m.Use(func(r render.Render) {
		// If API is stopping, render a HTTP 503
		if stopAPI {
			r.JSON(http.StatusServiceUnavailable, apiError{Message: "service is shutting down"})
			return
		}
	})

	// Set up API group routes
	r := martini.NewRouter()
	r.Group("/api", func(r martini.Router) {
		// Albums API
		r.Get("/albums", getAlbums)
		r.Get("/albums/:id", getAlbums)
	})

	// Add router action, start server
	// TODO: use port from configuration
	m.Action(r.Handle)
	go func() {
		if err := http.ListenAndServe(":8080", m); err != nil {
			log.Println(err)
		}
	}()

	// Trigger events via channel
	for {
		select {
		// Stop API
		case <-apiKillChan:
			// Stop serving requests
			stopAPI = true

			// Inform manager that shutdown is complete
			log.Println("api: stopped!")
			apiKillChan <- struct{}{}
			return
		}
	}
}

// apiError represents an error produced by the API
type apiError struct {
	Message string `json:"message"`
}

// apiAlbumsResponse represents the JSON response for /api/albums
type apiAlbumsResponse struct {
	Error  *apiError `json:"error"`
	Albums []Album   `json:"albums"`
	Songs  []Song    `json:"songs"`
}

// getAlbums retrieves one or more albums from wavepipe, and returns a HTTP status and JSON
func getAlbums(r render.Render, params martini.Params) {
	// Output struct for albums request
	res := apiAlbumsResponse{}

	// List of albums to return
	albums := make([]Album, 0)

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.Error = new(apiError)
			res.Error.Message = "invalid integer album ID"
			r.JSON(http.StatusBadRequest, res)
			return
		}

		// Load the album
		album := new(Album)
		album.ID = id
		if err := album.Load(); err != nil {
			res.Error = new(apiError)

			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.Error.Message = "album ID not found"
				r.JSON(http.StatusNotFound, res)
				return
			}

			// All other errors
			log.Println(err)
			res.Error.Message = "server error"
			r.JSON(http.StatusInternalServerError, res)
			return
		}

		// On single album, load the songs for this album
		songs, err := db.SongsForAlbum(album.ID)
		if err != nil {
			log.Println(err)
			res.Error = new(apiError)
			res.Error.Message = "server error"
			r.JSON(http.StatusInternalServerError, res)
			return
		}

		// Add songs to output
		res.Songs = songs

		// Add album to slice
		albums = append(albums, *album)
	} else {
		// Retrieve all albums
		tempAlbums, err := db.AllAlbums()
		if err != nil {
			log.Println(err)
			res.Error = new(apiError)
			res.Error.Message = "server error"
			r.JSON(http.StatusInternalServerError, res)
			return
		}

		// Copy albums into the output slice
		albums = tempAlbums
	}

	// Build response
	res.Error = nil
	res.Albums = albums

	// HTTP 200 OK with JSON
	r.JSON(http.StatusOK, res)
	return
}
