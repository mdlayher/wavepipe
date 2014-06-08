package core

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/api/auth"
	"github.com/mdlayher/wavepipe/config"
	"github.com/mdlayher/wavepipe/subsonic"

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
	m.Use(render.Renderer(render.Options{
		// Output human-readable JSON. GZIP will essentially negate the size increase, and this
		// makes the API much more developer-friendly
		IndentJSON: true,
	}))

	// Enable graceful shutdown when triggered by manager
	stopAPI := false
	m.Use(func(req *http.Request, res http.ResponseWriter, r render.Render) {
		// On debug, log everything
		if os.Getenv("WAVEPIPE_DEBUG") == "1" {
			log.Println(req.Header)
			log.Println(req.URL)
		}

		// Send a Server header with all responses
		res.Header().Set("Server", fmt.Sprintf("%s/%s (%s_%s)", App, Version, runtime.GOOS, runtime.GOARCH))

		// If API is stopping, render a HTTP 503
		if stopAPI {
			r.JSON(503, api.Error{
				Code:    503,
				Message: "service is shutting down",
			})
			return
		}
	})

	// Authenticate all API calls
	m.Use(func(req *http.Request, res http.ResponseWriter, c martini.Context, r render.Render) {
		// Use factory to determine the proper authentication method for this path
		method := auth.Factory(req.URL.Path)
		if method == nil {
			// If no method returned, path is not authenticated
			return
		}

		// Attempt authentication
		user, session, clientErr, serverErr := method.Authenticate(req)

		// Check for client error
		if clientErr != nil {
			// Check for a Subsonic error, since these are rendered as XML
			if subErr, ok := clientErr.(*subsonic.Container); ok {
				r.XML(200, subErr)
				return
			}

			// If debug mode, and no username or password, send a WWW-Authenticate header to prompt request
			// This allows for manual exploration of the API if needed
			if os.Getenv("WAVEPIPE_DEBUG") == "1" && (clientErr == auth.ErrNoUsername || clientErr == auth.ErrNoPassword) {
				res.Header().Set("WWW-Authenticate", "Basic")
			}

			r.JSON(401, api.Error{
				Code:    401,
				Message: "authentication failed: " + clientErr.Error(),
			})
			return
		}

		// Check for server error
		if serverErr != nil {
			log.Println(serverErr)

			// Check for a Subsonic error, since these are rendered as XML
			if subErr, ok := serverErr.(*subsonic.Container); ok {
				r.XML(200, subErr)
				return
			}

			r.JSON(500, api.Error{
				Code:    500,
				Message: "server error",
			})
			return
		}

		// Successful login, map session user and session to martini context
		c.Map(user)
		c.Map(session)

		// Print information about this API call
		log.Printf("api: [%s] %s", req.RemoteAddr, req.URL.Path)
	})

	// Set up API routes
	r := martini.NewRouter()

	// Set up API information route
	r.Get("/api", api.APIInfo)

	// Set up API group routes, with API version parameter
	r.Group("/api/:version", apiRoutes)

	// Set up emulated Subsonic API routes
	r.Group("/subsonic/rest", func(r martini.Router) {
		// Ping - used to check connectivity
		r.Get("/ping.view", subsonic.GetPing)

		// GetAlbumList2 - used to return a list of all albums by tags
		r.Get("/getAlbumList2.view", subsonic.GetAlbumList2)

		// GetAlbum - used to retrieve information about one album
		r.Get("/getAlbum.view", subsonic.GetAlbum)

		// GetRandomSongs - used to retrieve a number of random songs
		r.Get("/getRandomSongs.view", subsonic.GetRandomSongs)

		// Stream - used to return a binary file stream
		r.Get("/stream.view", subsonic.GetStream)
	})

	// On debug mode, enable pprof debug endpoints
	// Thanks: https://github.com/go-martini/martini/issues/228
	if os.Getenv("WAVEPIPE_DEBUG") == "1" {
		r.Group("/debug/pprof", func(r martini.Router) {
			r.Any("/", pprof.Index)
			r.Any("/cmdline", pprof.Cmdline)
			r.Any("/profile", pprof.Profile)
			r.Any("/symbol", pprof.Symbol)
			r.Any("/block", pprof.Handler("block").ServeHTTP)
			r.Any("/heap", pprof.Handler("heap").ServeHTTP)
			r.Any("/goroutine", pprof.Handler("goroutine").ServeHTTP)
			r.Any("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
		})
	}

	// Add router action, start server
	m.Action(r.Handle)
	go func() {
		// Load config
		conf, err := config.C.Load()
		if err != nil {
			log.Println(err)
			return
		}

		// Start server
		log.Println("api: listening on port", conf.Port)
		if err := http.ListenAndServe(":"+strconv.Itoa(conf.Port), m); err != nil {
			// Check if address in use
			if strings.Contains(err.Error(), "address already in use") {
				log.Fatalf("api: cannot bind to :%d, is wavepipe already running?", conf.Port)
			}

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

// apiRoutes sets up the API routes required by wavepipe
func apiRoutes(r martini.Router) {
	// Root API, containing information and help
	r.Get("", api.APIInfo)

	// Albums API
	r.Get("/albums", api.GetAlbums)
	r.Get("/albums/:id", api.GetAlbums)

	// Art API
	r.Get("/art", api.GetArt)
	r.Get("/art/:id", api.GetArt)

	// Artists API
	r.Get("/artists", api.GetArtists)
	r.Get("/artists/:id", api.GetArtists)

	// Folders API
	r.Get("/folders", api.GetFolders)
	r.Get("/folders/:id", api.GetFolders)

	// LastFM API
	r.Get("/lastfm", api.GetLastFM)
	r.Get("/lastfm/:action", api.GetLastFM)
	r.Get("/lastfm/:action/:id", api.GetLastFM)

	// Login API
	r.Get("/login", api.GetLogin)

	// Logout API
	r.Get("/logout", api.GetLogout)

	// Search API
	r.Get("/search", api.GetSearch)
	r.Get("/search/:query", api.GetSearch)

	// Songs API
	r.Get("/songs", api.GetSongs)
	r.Get("/songs/:id", api.GetSongs)

	// Status API
	r.Get("/status", api.GetStatus)

	// Stream API
	r.Get("/stream", api.GetStream)
	r.Get("/stream/:id", api.GetStream)

	// Transcode API
	r.Get("/transcode", api.GetTranscode)
	r.Get("/transcode/:id", api.GetTranscode)
}
