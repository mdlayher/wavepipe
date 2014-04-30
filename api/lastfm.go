package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/mdlayher/goset"
	"github.com/shkh/lastfm-go/lastfm"
)

const (
	// lfmAPIKey is the API key used to identify wavepipe to Last.fm
	lfmAPIKey = "f1dfe2dbca47ebc1f00adcb036b5de49"
	// lfmAPISecret is the API secret used to identify wavepipe to Last.fm
	lfmAPISecret = "8f68ccae06eda60e231418c881f5bfee"

	// lfmLogin is the action for a Last.fm login request
	lfmLogin = "login"
	// lfmNowPlaying is the action for a Last.fm now playing request
	lfmNowPlaying = "nowplaying"
	// lfmScrobble is the action for a Last.fm scrobble request
	lfmScrobble = "scrobble"
)

// LastFMResponse represents the JSON response for the Last.fm API
type LastFMResponse struct {
	Error  *Error        `json:"error"`
	URL    string        `json:"url"`
	render render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (l *LastFMResponse) RenderError(code int, message string) {
	// Nullify other fields
	l.URL = ""

	// Generate error
	l.Error = new(Error)
	l.Error.Code = code
	l.Error.Message = message

	// Render with specified HTTP status code
	l.render.JSON(code, l)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (l *LastFMResponse) ServerError() {
	l.RenderError(500, "server error")
	return
}

// GetLastFM allows access to the Last.fm API, enabling wavepipe to set a user's currently-playing
// track, as well as to enable scrobbling
func GetLastFM(req *http.Request, user *data.User, r render.Render, params martini.Params) {
	// Output struct for Last.fm response
	res := LastFMResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check API action
	action, ok := params["action"]
	if !ok {
		res.RenderError(400, "no string action provided")
		return
	}

	// Check for valid action
	if !set.New(lfmLogin, lfmNowPlaying, lfmScrobble).Has(action) {
		res.RenderError(400, "invalid string action provided")
	}

	// Instantiate Last.fm package
	lfm := lastfm.New(lfmAPIKey, lfmAPISecret)

	// Authenticate to the Last.fm API
	if action == lfmLogin {
		// Retrieve username from query
		username := req.URL.Query().Get("lfmu")
		if username == "" {
			res.RenderError(400, lfmLogin+": no username provided")
			return
		}

		// Retrieve password from query
		password := req.URL.Query().Get("lfmp")
		if password == "" {
			res.RenderError(400, lfmLogin+": no password provided")
			return
		}

		// Send a login request to Last.fm
		if err := lfm.Login(username, password); err != nil {
			res.RenderError(401, lfmLogin+": last.fm authentication failed")
			return
		}

		// Retrieve the API token for this user with wavepipe
		token, err := lfm.GetToken()
		if err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Store the user's Last.fm token in the database
		user.LastFMToken = token
		if err := user.Update(); err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Return the token authorization URL for the user
		res.URL = lfm.GetAuthTokenUrl(token)

		log.Println(lfmLogin, ": generated new token for user:", user.Username)

		// HTTP 200 OK with JSON
		res.Error = nil
		r.JSON(200, res)
		return
	}

	// All other actions require a valid token in database, and a valid integer song ID

	// Make sure this user has logged in using wavepipe before
	if user.LastFMToken == "" {
		res.RenderError(400, action+": user must authenticate to Last.fm")
		return
	}

	// Send a login request to Last.fm using token
	if err := lfm.LoginWithToken(user.LastFMToken); err != nil {
		// Check if token has not been authorized
		if strings.HasPrefix(err.Error(), "LastfmError[14]") {
			// Generate error output, but add the token authorization URL
			res = LastFMResponse{
				Error: &Error {
					Code: 401,
					Message: action+": last.fm token not yet authorized",
				},
				URL: lfm.GetAuthTokenUrl(user.LastFMToken),
			}

			// Output JSON
			r.JSON(401, res)
			return
		}

		// All other failures
		res.RenderError(401, action+": last.fm authentication failed")
		return
	}

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		res.RenderError(400, action+": no integer song ID provided")
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		res.RenderError(400, action+": invalid integer song ID")
		return
	}

	// Load the song by ID
	song := &data.Song{ID: id}
	if err := song.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			res.RenderError(404, action+": song ID not found")
			return
		}

		// All other errors
		log.Println(err)
		res.ServerError()
		return
	}

	// Log the current action
	log.Printf("%s : %s: %s - %s - %s", action, user.Username, song.Artist, song.Album, song.Title)

	// Create the track entity required by Last.fm from the song
	track := lastfm.P{
		"artist": song.Artist,
		"album": song.Album,
		"track": song.Title,
		"timestamp": time.Now().Unix(),
	}

	// Send a now playing request to the Last.fm API
	if action == lfmNowPlaying {
		// Perform the action
		if _, err := lfm.Track.UpdateNowPlaying(track); err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// HTTP 200 OK with JSON
		res.Error = nil
		r.JSON(200, res)
		return
	}

	// Send a scrobble request to the Last.fm API
	if action == lfmScrobble {
		// Perform the action
		if _, err := lfm.Track.Scrobble(track); err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// HTTP 200 OK with JSON
		res.Error = nil
		r.JSON(200, res)
		return
	}

	// Invalid action, meaning programmer error, HTTP 500
	log.Println("no such Last.fm action:", action)
	res.ServerError()
	return
}
