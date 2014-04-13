package core

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/api"

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
			r.JSON(http.StatusServiceUnavailable, api.Error{Message: "service is shutting down"})
			return
		}
	})

	// Set up API group routes
	r := martini.NewRouter()
	r.Group("/api", func(r martini.Router) {
		// Albums API
		r.Get("/albums", api.GetAlbums)
		r.Get("/albums/:id", api.GetAlbums)

		// Artists API
		r.Get("/artists", api.GetArtists)
		r.Get("/artists/:id", api.GetArtists)

		// Songs API
		r.Get("/songs", api.GetSongs)
		r.Get("/songs/:id", api.GetSongs)
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
