package core

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/api/auth"
	"github.com/mdlayher/wavepipe/config"

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
	m.Use(func(res http.ResponseWriter, r render.Render) {
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
		// Set a different authentication method depending on endpoint
		var authMethod auth.AuthMethod

		// Enable simple authentication for local clients, if special password set
		path := strings.TrimRight(req.URL.Path, "/")
		if strings.Contains(req.RemoteAddr, "127.0.0.1") && req.URL.Query().Get("p") == "simple" {
			// Use simple authentication
			authMethod = new(auth.SimpleAuth)
		} else if path == "/api/" + api.APIVersion + "/login" {
			// For login, use the bcrypt authenticator to generate a new session
			authMethod = new(auth.BcryptAuth)
		} else if len(path) > 3 && path[0:4] == "/api" {
			// For other API methods, use the HMAC-SHA1 authenticator
			authMethod = new(auth.HMACAuth)
		} else {
			// For any other endpoints, no authentication
			return
		}

		// Attempt authentication
		user, session, clientErr, serverErr := authMethod.Authenticate(req)

		// Check for client error
		if clientErr != nil {
			// If no username or password, send a WWW-Authenticate header to prompt request
			// This allows for manual exploration of the API if needed
			if clientErr == auth.ErrNoUsername || clientErr == auth.ErrNoPassword {
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
	r.Group("/api/:version", func(r martini.Router) {
		// Root API, containing information and help
		r.Get("", api.APIInfo)

		// Albums API
		r.Get("/albums", api.GetAlbums)
		r.Get("/albums/:id", api.GetAlbums)

		// Artists API
		r.Get("/artists", api.GetArtists)
		r.Get("/artists/:id", api.GetArtists)

		// Folders API
		r.Get("/folders", api.GetFolders)
		r.Get("/folders/:id", api.GetFolders)

		// Login API
		r.Get("/login", api.GetLogin)

		// Logout API
		r.Get("/logout", api.GetLogout)

		// Songs API
		r.Get("/songs", api.GetSongs)
		r.Get("/songs/:id", api.GetSongs)

		// Stream API
		r.Get("/stream", api.GetStream)
		r.Get("/stream/:id", api.GetStream)
	})

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
