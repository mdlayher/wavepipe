package core

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// Martini instance to test against
var m = martini.New()

func init() {
	// Set up database connection
	data.DB = new(data.SqliteBackend)
	data.DB.DSN("~/.config/wavepipe/wavepipe.db")
	if err := data.DB.Open(); err != nil {
		os.Exit(1)
	}

	// Map a dummy user for some API calls
	m.Use(render.Renderer(render.Options{}))
	m.Use(func(c martini.Context) {
		c.Map(new(data.User))
	})

	// Set up Martini with API routes
	r := martini.NewRouter()
	r.Group("/api/:version", apiRoutes)
	m.Action(r.Handle)

}

// TestAPIRouter verifies that all API request processing functionality is working properly
func TestAPIRouter(t *testing.T) {
	// Table of tests to run, and their expected HTTP status results
	var tests = []struct {
		code int
		url  string
	}{
		// Albums API
		//   - valid request
		{200, "/api/v0/albums"},
		//   - valid request for 1 item
		{200, "/api/v0/albums/1"},
		//   - valid limit items request
		{200, "/api/v0/albums?limit=0,10"},
		//   - invalid API version
		{400, "/api/v999/albums"},
		//   - invalid integer album ID
		{400, "/api/v0/albums/foo"},
		//   - missing second integer for limit
		{400, "/api/v0/albums?limit=0"},
		//   - invalid integer pair for limit
		{400, "/api/v0/albums?limit=foo,bar"},
		//   - album ID not found
		{404, "/api/v0/albums/99999999"},

		// Art API - skip valid requests, due to binary output
		//   - invalid API version
		{400, "/api/v999/art"},
		//   - no integer ID provided
		{400, "/api/v0/art"},
		//   - invalid art stream ID
		{400, "/api/v0/art/foo"},
		//   - art ID not found
		{404, "/api/v0/art/99999999"},

		// Artists API
		//   - valid request
		{200, "/api/v0/artists"},
		//   - valid request for 1 item
		{200, "/api/v0/artists/1"},
		//   - invalid API version
		{400, "/api/v999/artists"},
		//   - invalid integer artist ID
		{400, "/api/v0/artists/foo"},
		//   - missing second integer for limit
		{400, "/api/v0/artists?limit=0"},
		//   - invalid integer pair for limit
		{400, "/api/v0/artists?limit=foo,bar"},
		//   - artist ID not found
		{404, "/api/v0/artists/99999999"},

		// Folders API
		//   - valid request
		{200, "/api/v0/folders"},
		//   - valid request for 1 item
		{200, "/api/v0/folders/1"},
		//   - invalid API version
		{400, "/api/v999/folders"},
		//   - invalid integer folder ID
		{400, "/api/v0/folders/foo"},
		//   - missing second integer for limit
		{400, "/api/v0/folders?limit=0"},
		//   - invalid integer pair for limit
		{400, "/api/v0/folders?limit=foo,bar"},
		//   - folder ID not found
		{404, "/api/v0/folders/99999999"},

		// LastFM API - skip valid requests, due to need for external service
		//   - invalid API version
		{400, "/api/v999/lastfm"},
		//   - no string action provided
		{400, "/api/v0/lastfm"},
		//   - invalid string action provided
		{400, "/api/v0/lastfm/foo"},
		//   - login: no username provided
		{400, "/api/v0/lastfm/login"},
		//   - login: no password provided
		{400, "/api/v0/lastfm/login?lfmu=foo"},
		//   - action: user must authenticate to last.fm
		{401, "/api/v0/lastfm/nowplaying"},
		//   - action: user must authenticate to last.fm
		{401, "/api/v0/lastfm/scrobble"},
		// Cannot test other calls without a valid Last.fm token

		// Login/Logout API - skip due to need for sessions and users

		// Search API
		//   - valid request
		{200, "/api/v0/search/foo"},
		//   - invalid API version
		{400, "/api/v999/search"},
		//   - no search query specified
		{400, "/api/v0/search"},

		// Songs API
		//   - valid request
		{200, "/api/v0/songs"},
		//   - valid request for 1 item
		{200, "/api/v0/songs/1"},
		//   - invalid API version
		{400, "/api/v999/songs"},
		//   - invalid integer song ID
		{400, "/api/v0/songs/foo"},
		//   - missing second integer for limit
		{400, "/api/v0/songs?limit=0"},
		//   - invalid integer pair for limit
		{400, "/api/v0/songs?limit=foo,bar"},
		//   - song ID not found
		{404, "/api/v0/songs/99999999"},

		// Status API
		//   - valid request
		{200, "/api/v0/status"},
		//   - invalid API version
		{400, "/api/v999/status"},

		// Stream API - skip valid requests, due to binary output
		//   - invalid API version
		{400, "/api/v999/stream"},
		//   - no integer stream ID provided
		{400, "/api/v0/stream"},
		//   - invalid stream stream ID
		{400, "/api/v0/stream/foo"},
		//   - song ID not found
		{404, "/api/v0/stream/99999999"},

		// Transcode API - skip valid requests, due to binary output
		//   - invalid API version
		{400, "/api/v999/transcode"},
		//   - no integer transcode ID provided
		{400, "/api/v0/transcode"},
		//   - invalid transcode transcode ID
		{400, "/api/v0/transcode/foo"},
		//   - song ID not found
		{404, "/api/v0/transcode/99999999"},
		//   - ffmpeg not found, transcoding disabled
		{503, "/api/v0/transcode/1"},
	}

	// Iterate all tests
	for _, test := range tests {
		// Generate a new HTTP request
		r, err := http.NewRequest("GET", "http://localhost:8080"+test.url, nil)
		if err != nil {
			t.Fatalf("Failed to create HTTP request")
		}

		// Capture HTTP response via recorder
		w := httptest.NewRecorder()

		// Perform request
		m.ServeHTTP(w, r)

		// Validate results
		if w.Code != test.code {
			t.Fatalf("[%v != %v] %s", w.Code, test.code, test.url)
		}

		log.Printf("OK: [%d] %s", test.code, test.url)
	}
}
