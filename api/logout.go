package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// GetLogout destroys an existing session from the wavepipe API, and returns a HTTP status and JSON
func GetLogout(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Attempt to retrieve session from context
	session := new(data.Session)
	if tempSession := context.Get(req, CtxSession); tempSession != nil {
		session = tempSession.(*data.Session)
	} else {
		// No session stored in context
		log.Println("api: no session stored in request context!")
		r.JSON(res, 500, serverErr)
		return
	}

	// Output struct for logout request
	out := ErrorResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Destroy the current API session
	if session != nil {
		if err := session.Delete(); err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}
	}

	// No errors
	out.Error = nil

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
