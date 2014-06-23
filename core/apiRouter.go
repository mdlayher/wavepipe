package core

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/api/auth"
	"github.com/mdlayher/wavepipe/config"
	"github.com/mdlayher/wavepipe/data"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/stretchr/graceful"
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

	// GZIP all responses
	n.Use(gzip.Gzip(gzip.DefaultCompression))

	// Initial API setup
	n.Use(negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		// On debug, log everything
		if os.Getenv("WAVEPIPE_DEBUG") == "1" {
			log.Println(req.Header)
			log.Println(req.URL)
		}

		// Send a Server header with all responses
		res.Header().Set("Server", fmt.Sprintf("%s/%s (%s_%s)", App, Version, runtime.GOOS, runtime.GOARCH))

		// Store render in context for all API calls
		context.Set(req, api.CtxRender, r)

		// Delegate to next middleware
		next(res, req)
	}))

	// Authenticate all API calls
	n.Use(negroni.HandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		// Use factory to determine and invoke the proper authentication method for this path
		user, session, clientErr, serverErr := auth.Factory(req.URL.Path).Authenticate(req)

		// Check for client error
		if clientErr != nil {
			// If debug mode, and no username or password, send a WWW-Authenticate header to prompt request
			// This allows for manual exploration of the API if needed
			if os.Getenv("WAVEPIPE_DEBUG") == "1" && (clientErr == auth.ErrNoUsername || clientErr == auth.ErrNoPassword) {
				res.Header().Set("WWW-Authenticate", "Basic")
			}

			r.JSON(res, 401, api.ErrorResponse{&api.Error{
				Code:    401,
				Message: "authentication failed: " + clientErr.Error(),
			}})
			return
		}

		// Check for server error
		if serverErr != nil {
			log.Println(serverErr)

			r.JSON(res, 500, api.ErrorResponse{&api.Error{
				Code:    500,
				Message: "server error",
			}})
			return
		}

		// Successful login, map session user and session to gorilla context for this request
		context.Set(req, api.CtxUser, user)
		context.Set(req, api.CtxSession, session)

		// Print information about this API call
		log.Printf("api: [%s] %s %s?%s", req.RemoteAddr, req.Method, req.URL.Path, req.URL.Query().Encode())

		// Perform API call
		next(res, req)
	}))

	// Wait for graceful to signal termination
	gracefulChan := make(chan struct{}, 0)

	// Use gorilla mux with negroni, start server
	n.UseHandler(newRouter())
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

		// Start server, allowing up to 10 seconds after shutdown for clients to complete
		log.Println("api: binding to host", conf.Host)
		if err := graceful.ListenAndServe(&http.Server{Addr: conf.Host, Handler: n}, 10*time.Second); err != nil {
			// Check if address in use
			if strings.Contains(err.Error(), "address already in use") {
				log.Fatalf("api: cannot bind to %s, is wavepipe already running?", conf.Host)
			}

			// Ignore error on closing
			if !strings.Contains(err.Error(), "use of closed network connection") {
				// Log other errors
				log.Println(err)
			}
		}

		// Shutdown complete
		close(gracefulChan)
	}()

	// Trigger events via channel
	for {
		select {
		// Stop API
		case <-apiKillChan:
			// If testing, don't wait for graceful shutdown
			if os.Getenv("WAVEPIPE_TEST") != "1" {
				// Block and wait for graceful shutdown
				log.Println("api: waiting for remaining connections to close...")
				<-gracefulChan
			}

			// Inform manager that shutdown is complete
			log.Println("api: stopped!")
			apiKillChan <- struct{}{}
			return
		}
	}
}

// newRouter sets up the web and API routes required by wavepipe
func newRouter() *mux.Router {
	// Create a router
	router := mux.NewRouter().StrictSlash(false)

	// HTTP handler for web UI
	webUI := func(res http.ResponseWriter, req *http.Request) {
		// Retrieve render
		r := context.Get(req, api.CtxRender).(*render.Render)

		// Get the asset name
		name := mux.Vars(req)["asset"]

		// If asset name empty, return the index
		if name == "" {
			name = "index.html"
		}

		// More information on debug
		if os.Getenv("WAVEPIPE_DEBUG") == "1" {
			log.Println("web: fetching resource: res/web/" + name)
		}

		// Retrieve asset
		asset, err := data.Asset("res/web/" + name)
		if err != nil {
			res.WriteHeader(404)
			return
		}

		// Render asset and return its type
		res.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(name)))
		r.Data(res, 200, asset)
	}

	// Web UI and its assets
	router.HandleFunc("/", webUI).Methods("GET")
	router.HandleFunc("/res/{asset:.*}", webUI).Methods("GET")

	// Set up robots.txt to disallow crawling, since this is a dynamic service which users self-host
	router.HandleFunc("/robots.txt", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("# wavepipe media server\n" +
			"# https://github.com/mdlayher/wavepipe\n" +
			"User-agent: *\n" +
			"Disallow: /"))
	}).Methods("GET")

	// Set up API information route
	router.HandleFunc("/api", api.APIInfo).Methods("GET")

	// Set up API group routes, with API version parameter
	ar := router.PathPrefix("/api/{version}/").Subrouter()

	// Albums API
	ar.HandleFunc("/albums", api.GetAlbums).Methods("GET")
	ar.HandleFunc("/albums/{id}", api.GetAlbums).Methods("GET")

	// Art API
	ar.HandleFunc("/art", api.GetArt).Methods("GET")
	ar.HandleFunc("/art/{id}", api.GetArt).Methods("GET")

	// Artists API
	ar.HandleFunc("/artists", api.GetArtists).Methods("GET")
	ar.HandleFunc("/artists/{id}", api.GetArtists).Methods("GET")

	// Folders API
	ar.HandleFunc("/folders", api.GetFolders).Methods("GET")
	ar.HandleFunc("/folders/{id}", api.GetFolders).Methods("GET")

	// LastFM API
	ar.HandleFunc("/lastfm", api.PostLastFM).Methods("POST")
	ar.HandleFunc("/lastfm/{action}", api.PostLastFM).Methods("POST")
	ar.HandleFunc("/lastfm/{action}/{id}", api.PostLastFM).Methods("POST")

	// Login API
	ar.HandleFunc("/login", api.PostLogin).Methods("POST")

	// Logout API
	ar.HandleFunc("/logout", api.PostLogout).Methods("POST")

	// Search API
	ar.HandleFunc("/search", api.GetSearch).Methods("GET")
	ar.HandleFunc("/search/{query}", api.GetSearch).Methods("GET")

	// Songs API
	ar.HandleFunc("/songs", api.GetSongs).Methods("GET")
	ar.HandleFunc("/songs/{id}", api.GetSongs).Methods("GET")

	// Status API
	ar.HandleFunc("/status", api.GetStatus).Methods("GET")

	// Stream API
	ar.HandleFunc("/stream", api.GetStream).Methods("GET")
	ar.HandleFunc("/stream/{id}", api.GetStream).Methods("GET")

	// Transcode API
	ar.HandleFunc("/transcode", api.GetTranscode).Methods("GET")
	ar.HandleFunc("/transcode/{id}", api.GetTranscode).Methods("GET")

	// Users API
	ar.HandleFunc("/users", api.GetUsers).Methods("GET")
	ar.HandleFunc("/users/{id}", api.GetUsers).Methods("GET")
	ar.HandleFunc("/users", api.PostUsers).Methods("POST")
	ar.HandleFunc("/users/{id}", api.PutUsers).Methods("PUT", "PATCH")
	ar.HandleFunc("/users/{id}", api.DeleteUsers).Methods("DELETE")

	// On debug mode, enable pprof debug endpoints
	// Thanks: https://github.com/go-martini/martini/issues/228
	if os.Getenv("WAVEPIPE_DEBUG") == "1" {
		dr := router.PathPrefix("/debug/pprof").Subrouter()
		dr.HandleFunc("/", pprof.Index)
		dr.HandleFunc("/cmdline", pprof.Cmdline)
		dr.HandleFunc("/profile", pprof.Profile)
		dr.HandleFunc("/symbol", pprof.Symbol)
		dr.HandleFunc("/block", pprof.Handler("block").ServeHTTP)
		dr.HandleFunc("/heap", pprof.Handler("heap").ServeHTTP)
		dr.HandleFunc("/goroutine", pprof.Handler("goroutine").ServeHTTP)
		dr.HandleFunc("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	}

	// Return configured router
	return router
}
