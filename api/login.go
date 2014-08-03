package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// LoginResponse represents the JSON response for /api/logins
type LoginResponse struct {
	Error   *Error        `json:"error"`
	Session *data.Session `json:"session"`
}

// PostLogin creates a new session for the input user to use the wavepipe API,
// and returns a HTTP status and JSON.
func PostLogin(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Attempt to retrieve user from context
	user := new(data.User)
	if tempUser := context.Get(r, CtxUser); tempUser != nil {
		user = tempUser.(*data.User)
	} else {
		// No user stored in context
		log.Println("api: no user stored in request context!")
		ren.JSON(w, 500, serverErr)
		return
	}

	// Output struct for login request
	out := LoginResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Generate a new API session for this user, with optional specified session name
	// via "client" POST parameter
	session, err := user.CreateSession(r.PostFormValue("client"))
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Session = session
	ren.JSON(w, 200, out)
	return
}
