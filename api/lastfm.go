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
	"github.com/mdlayher/goset"
	"github.com/mdlayher/render"
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
	Error *Error `json:"error"`
	URL   string `json:"url"`
}

// GetLastFM allows access to the Last.fm API, enabling wavepipe to set a user's currently-playing
// track, as well as to enable scrobbling
func GetLastFM(req *http.Request, user *data.User, r render.Render, params martini.Params) {
	// Output struct for Last.fm response
	res := LastFMResponse{}

	// Output struct for errors
	errRes := ErrorResponse{render: r}

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			errRes.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check API action
	action, ok := params["action"]
	if !ok {
		errRes.RenderError(400, "no string action provided")
		return
	}

	// Check for valid action
	if !set.New(lfmLogin, lfmNowPlaying, lfmScrobble).Has(action) {
		errRes.RenderError(400, "invalid string action provided")
	}

	// Instantiate Last.fm package
	lfm := lastfm.New(lfmAPIKey, lfmAPISecret)

	// Authenticate to the Last.fm API
	if action == lfmLogin {
		// Retrieve username from query
		username := req.URL.Query().Get("lfmu")
		if username == "" {
			errRes.RenderError(400, lfmLogin+": no username provided")
			return
		}

		// Retrieve password from query
		password := req.URL.Query().Get("lfmp")
		if password == "" {
			errRes.RenderError(400, lfmLogin+": no password provided")
			return
		}

		// Send a login request to Last.fm
		if err := lfm.Login(username, password); err != nil {
			errRes.RenderError(401, lfmLogin+": last.fm authentication failed")
			return
		}

		// Retrieve the API token for this user with wavepipe
		token, err := lfm.GetToken()
		if err != nil {
			log.Println(err)
			errRes.ServerError()
			return
		}

		// Store the user's Last.fm token in the database
		user.LastFMToken = token
		if err := user.Update(); err != nil {
			log.Println(err)
			errRes.ServerError()
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
		errRes.RenderError(401, action+": user must authenticate to last.fm")
		return
	}

	// Send a login request to Last.fm using token
	if err := lfm.LoginWithToken(user.LastFMToken); err != nil {
		// Check if token has not been authorized
		if strings.HasPrefix(err.Error(), "LastfmError[14]") {
			// Generate error output, but add the token authorization URL
			res.URL = lfm.GetAuthTokenUrl(user.LastFMToken)
			errRes.RenderError(401, action+": last.fm token not yet authorized")
			return
		}

		// All other failures
		errRes.RenderError(401, action+": last.fm authentication failed")
		return
	}

	// Check for an ID parameter
	pID, ok := params["id"]
	if !ok {
		errRes.RenderError(400, action+": no integer song ID provided")
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		errRes.RenderError(400, action+": invalid integer song ID")
		return
	}

	// Load the song by ID
	song := &data.Song{ID: id}
	if err := song.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			errRes.RenderError(404, action+": song ID not found")
			return
		}

		// All other errors
		log.Println(err)
		errRes.ServerError()
		return
	}

	// Log the current action
	log.Printf("%s : %s : [#%05d] %s - %s", action, user.Username, song.ID, song.Artist, song.Title)

	// Create the track entity required by Last.fm from the song
	track := lastfm.P{
		"artist":    song.Artist,
		"album":     song.Album,
		"track":     song.Title,
		"timestamp": time.Now().Unix(),
	}

	// Check for optional timestamp parameter, which could be useful for sending scrobbles at
	// past times, etc
	if pTS, ok := params["timestamp"]; ok {
		// Verify valid integer timestamp
		ts, err := strconv.Atoi(pTS)
		if err != nil || ts < 0 {
			errRes.RenderError(400, action+": invalid integer timestamp")
			return
		}

		// Override previously set timestamp with this one
		track["timestamp"] = ts
	}

	// Send a now playing request to the Last.fm API
	if action == lfmNowPlaying {
		// Perform the action
		if _, err := lfm.Track.UpdateNowPlaying(track); err != nil {
			log.Println(err)
			errRes.ServerError()
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
			errRes.ServerError()
			return
		}

		// HTTP 200 OK with JSON
		res.Error = nil
		r.JSON(200, res)
		return
	}

	// Invalid action, meaning programmer error, HTTP 500
	log.Println("no such Last.fm action:", action)
	errRes.ServerError()
	return
}
