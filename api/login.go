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

// PostLogin creates a new session on the wavepipe API, and returns a HTTP status and JSON
func PostLogin(res http.ResponseWriter, req *http.Request) {
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

	// Output struct for login request
	out := LoginResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Generate a new API session for this user, with optional specified session name
	// via "client" POST parameter
	session, err := user.CreateSession(req.PostFormValue("client"))
	if err != nil {
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Build response
	out.Error = nil
	out.Session = session

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
