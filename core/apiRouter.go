package core

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/api/auth"
	"github.com/mdlayher/wavepipe/config"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// apiRouter sets up the instance of negroni
func apiRouter(apiKillChan chan struct{}) {
	log.Println("api: starting...")

	// Initialize negroni
	n := negroni.New()

	// Set up render
	r := render.New(render.Options{
		// Output human-readable JSON. GZIP will essentially negate the size increase, and this
		// makes the API much more developer-friendly
		IndentJSON: true,
	})

	// Enable graceful shutdown when triggered by manager
	stopAPI := false
	n.Use(negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		// On debug, log everything
		if os.Getenv("WAVEPIPE_DEBUG") == "1" {
			log.Println(req.Header)
			log.Println(req.URL)
		}

		// Send a Server header with all responses
		res.Header().Set("Server", fmt.Sprintf("%s/%s (%s_%s)", App, Version, runtime.GOOS, runtime.GOARCH))

		// If API is stopping, render a HTTP 503
		if stopAPI {
			r.JSON(res, 503, api.Error{
				Code:    503,
				Message: "service is shutting down",
			})
			return
		}

		// Store render in context for all API calls
		context.Set(req, api.CtxRender, r)

		// Delegate to next middleware
		next(res, req)
	}))

	// Authenticate all API calls
	n.Use(negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		// Delegate to next middleware on any return
		defer next(res, req)

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
			// If debug mode, and no username or password, send a WWW-Authenticate header to prompt request
			// This allows for manual exploration of the API if needed
			if os.Getenv("WAVEPIPE_DEBUG") == "1" && (clientErr == auth.ErrNoUsername || clientErr == auth.ErrNoPassword) {
				res.Header().Set("WWW-Authenticate", "Basic")
			}

			r.JSON(res, 401, api.Error{
				Code:    401,
				Message: "authentication failed: " + clientErr.Error(),
			})
			return
		}

		// Check for server error
		if serverErr != nil {
			log.Println(serverErr)

			r.JSON(res, 500, api.Error{
				Code:    500,
				Message: "server error",
			})
			return
		}

		// Successful login, map session user and session to gorilla context for this request
		context.Set(req, api.CtxUser, user)
		context.Set(req, api.CtxSession, session)

		// Print information about this API call
		log.Printf("api: [%s] %s?%s", req.RemoteAddr, req.URL.Path, req.URL.Query().Encode())
	}))

	// Set up API routes
	router := mux.NewRouter().StrictSlash(false)

	// Set up robots.txt to disallow crawling, since this is a dynamic service which users self-host
	router.HandleFunc("/robots.txt", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("# wavepipe media server\n" +
			"# https://github.com/mdlayher/wavepipe\n" +
			"User-agent: *\n" +
			"Disallow: /"))
	}).Methods("GET")

	// Set up API information route
	router.HandleFunc("/api/", api.APIInfo).Methods("GET")

	// Set up API group routes, with API version parameter
	subrouter := router.PathPrefix("/api/{version}/").Subrouter()
	apiRoutes(subrouter)

	// On debug mode, enable pprof debug endpoints
	/*
		// Thanks: https://github.com/go-negroni/negroni/issues/228
		if os.Getenv("WAVEPIPE_DEBUG") == "1" {
			r.Group("/debug/pprof", func(r negroni.Router) {
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
	*/

	// Use gorilla mux with negroni, start server
	n.UseHandler(router)
	go func() {
		// Load config
		conf, err := config.C.Load()
		if err != nil {
			log.Println(err)
			return
		}

		// Check for empty host
		if conf.Host == "" {
			log.Fatalf("api: no host specified in configuration")
		}

		// Start server
		log.Println("api: binding to host", conf.Host)
		if err := http.ListenAndServe(conf.Host, n); err != nil {
			// Check if address in use
			if strings.Contains(err.Error(), "address already in use") {
				log.Fatalf("api: cannot bind to %s, is wavepipe already running?", conf.Host)
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
func apiRoutes(r *mux.Router) {
	// Albums API
	r.HandleFunc("/albums", api.GetAlbums).Methods("GET")
	r.HandleFunc("/albums/{id}", api.GetAlbums).Methods("GET")

	// Art API
	r.HandleFunc("/art", api.GetArt).Methods("GET")
	r.HandleFunc("/art/{id}", api.GetArt).Methods("GET")

	// Artists API
	r.HandleFunc("/artists", api.GetArtists).Methods("GET")
	r.HandleFunc("/artists/{id}", api.GetArtists).Methods("GET")

	// Folders API
	r.HandleFunc("/folders", api.GetFolders).Methods("GET")
	r.HandleFunc("/folders/{id}", api.GetFolders).Methods("GET")

	// LastFM API
	r.HandleFunc("/lastfm", api.GetLastFM).Methods("GET")
	r.HandleFunc("/lastfm/{action}", api.GetLastFM).Methods("GET")
	r.HandleFunc("/lastfm/{action}/{id}", api.GetLastFM).Methods("GET")

	// Login API
	r.HandleFunc("/login", api.GetLogin).Methods("GET")

	/*

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

	*/

	// Stream API
	r.HandleFunc("/stream", api.GetStream).Methods("GET")
	r.HandleFunc("/stream/{id}", api.GetStream).Methods("GET")

	// Transcode API
	r.HandleFunc("/transcode", api.GetTranscode).Methods("GET")
	r.HandleFunc("/transcode/{id}", api.GetTranscode).Methods("GET")
}
