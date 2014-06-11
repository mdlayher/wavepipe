package core

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// Negroni instance to test against
var n = negroni.New()

// Render instance to test against
var r = render.New(render.Options{})

func init() {
	// Set up database connection
	data.DB = new(data.SqliteBackend)
	data.DB.DSN("~/.config/wavepipe/wavepipe.db")
	if err := data.DB.Open(); err != nil {
		os.Exit(1)
	}

	// Set up Negroni with API routes
	router := mux.NewRouter().StrictSlash(false)
	subrouter := router.PathPrefix("/api/{version}/").Subrouter()
	apiRoutes(subrouter)
	n.UseHandler(router)
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
		//   - valid limit items request
		{200, "/api/v0/artists?limit=0,10"},
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
		//   - valid limit items request
		{200, "/api/v0/folders?limit=0,10"},
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
		//   - valid limit items request
		{200, "/api/v0/songs?limit=0,10"},
		//   - valid random items request
		{200, "/api/v0/songs?random=10"},
		//   - invalid API version
		{400, "/api/v999/songs"},
		//   - invalid integer song ID
		{400, "/api/v0/songs/foo"},
		//   - missing second integer for limit
		{400, "/api/v0/songs?limit=0"},
		//   - invalid integer pair for limit
		{400, "/api/v0/songs?limit=foo,bar"},
		//   - invalid integer for random
		{400, "/api/v0/songs?random=foo"},
		//   - song ID not found
		{404, "/api/v0/songs/99999999"},

		// Status API
		//   - valid request
		{200, "/api/v0/status"},
		//   - valid request with metrics
		{200, "/api/v0/status?metrics=true"},
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
		req, err := http.NewRequest("GET", "http://localhost:8080"+test.url, nil)
		if err != nil {
			t.Fatalf("Failed to create HTTP request")
		}

		// Map context for request
		context.Set(req, api.CtxRender, r)
		context.Set(req, api.CtxUser, new(data.User))
		context.Set(req, api.CtxSession, new(data.Session))

		// Capture HTTP response via recorder
		w := httptest.NewRecorder()

		// Perform request
		n.ServeHTTP(w, req)

		// Validate results
		if w.Code != test.code {
			t.Fatalf("HTTP [%v != %v] %s", w.Code, test.code, test.url)
		}

		// Check result body as well
		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Unmarshal error response
		var errRes api.ErrorResponse
		if err := json.Unmarshal(body, &errRes); err != nil {
			log.Println(string(body))
			t.Fatal(err)
		}

		// If not HTTP 200, check to ensure error code matches
		if errRes.Error != nil && errRes.Error.Code != test.code {
			t.Fatalf("Body [%v != %v] %s", w.Code, test.code, test.url)
		}

		log.Printf("OK: [%d] %s", test.code, test.url)
	}
}
