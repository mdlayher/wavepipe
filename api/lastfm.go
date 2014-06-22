package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/goset"
	"github.com/shkh/lastfm-go/lastfm"
	"github.com/unrolled/render"
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

// PostLastFM allows access to the Last.fm API, enabling wavepipe to set a user's currently-playing
// track, as well as to enable scrobbling
func PostLastFM(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Attempt to retrieve user from context
	user := new(data.User)
	if tempUser := context.Get(req, CtxUser); tempUser != nil {
		user = tempUser.(*data.User)
	} else {
		// No user stored in context
		log.Println("api: no user stored in request context!")
		r.JSON(res, 500, serverErr)
		return
	}

	// Output struct for Last.fm response
	out := LastFMResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Do not allow guests and below to use Last.fm functionality
	if user.RoleID < data.RoleUser {
		r.JSON(res, 403, permissionErr)
		return
	}

	// Check API action
	action, ok := mux.Vars(req)["action"]
	if !ok {
		r.JSON(res, 400, errRes(400, "no string action provided"))
		return
	}

	// Check for valid action
	if !set.New(lfmLogin, lfmNowPlaying, lfmScrobble).Has(action) {
		r.JSON(res, 400, errRes(400, "invalid string action provided"))
		return
	}

	// Instantiate Last.fm package
	lfm := lastfm.New(lfmAPIKey, lfmAPISecret)

	// Authenticate to the Last.fm API
	if action == lfmLogin {
		// Retrieve username from POST body
		username := req.PostFormValue("username")
		if username == "" {
			r.JSON(res, 400, errRes(400, lfmLogin+": no username provided"))
			return
		}

		// Retrieve password from POST body
		password := req.PostFormValue("password")
		if password == "" {
			r.JSON(res, 400, errRes(400, lfmLogin+": no password provided"))
			return
		}

		// Send a login request to Last.fm
		if err := lfm.Login(username, password); err != nil {
			r.JSON(res, 401, errRes(401, lfmLogin+": last.fm authentication failed"))
			return
		}

		// Retrieve the API token for this user with wavepipe
		token, err := lfm.GetToken()
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Store the user's Last.fm token in the database
		user.LastFMToken = token
		if err := user.Update(); err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Return the token authorization URL for the user
		out.URL = lfm.GetAuthTokenUrl(token)

		log.Println(lfmLogin, ": generated new token for user:", user.Username)

		// HTTP 200 OK with JSON
		out.Error = nil
		r.JSON(res, 200, out)
		return
	}

	// All other actions require a valid token in database, and a valid integer song ID

	// Make sure this user has logged in using wavepipe before
	if user.LastFMToken == "" {
		r.JSON(res, 401, errRes(401, action+": user must authenticate to last.fm"))
		return
	}

	// Send a login request to Last.fm using token
	if err := lfm.LoginWithToken(user.LastFMToken); err != nil {
		// Check if token has not been authorized
		if strings.HasPrefix(err.Error(), "LastfmError[14]") {
			// Generate error output, but add the token authorization URL
			out.URL = lfm.GetAuthTokenUrl(user.LastFMToken)
			r.JSON(res, 401, errRes(401, action+": last.fm token not yet authorized"))
			return
		}

		// All other failures
		r.JSON(res, 401, errRes(401, action+": last.fm authentication failed"))
		return
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(req)["id"]
	if !ok {
		r.JSON(res, 400, errRes(400, action+": no integer song ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		r.JSON(res, 400, errRes(400, action+": invalid integer song ID"))
		return
	}

	// Load the song by ID
	song := &data.Song{ID: id}
	if err := song.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			r.JSON(res, 404, errRes(404, action+": song ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		r.JSON(res, 500, serverErr)
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
	if pTS := req.URL.Query().Get("timestamp"); pTS != "" {
		// Verify valid integer timestamp
		ts, err := strconv.Atoi(pTS)
		if err != nil || ts < 0 {
			r.JSON(res, 400, errRes(400, action+": invalid integer timestamp"))
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
			r.JSON(res, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Error = nil
		r.JSON(res, 200, out)
		return
	}

	// Send a scrobble request to the Last.fm API
	if action == lfmScrobble {
		// Perform the action
		if _, err := lfm.Track.Scrobble(track); err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Error = nil
		r.JSON(res, 200, out)
		return
	}

	// Invalid action, meaning programmer error, HTTP 500
	log.Println("no such Last.fm action:", action)
	r.JSON(res, 500, serverErr)
	return
}
