package api

import (
	"log"
	"net/http"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// PostLogout destroys an existing session from the wavepipe API,
// and returns a HTTP status and JSON.
func PostLogout(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Attempt to retrieve session from context
	session := new(data.Session)
	if tempSession := context.Get(r, CtxSession); tempSession != nil {
		session = tempSession.(*data.Session)
	} else {
		// No session stored in context
		log.Println("api: no session stored in request context!")
		ren.JSON(w, 500, serverErr)
		return
	}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Destroy the current API session
	if session != nil {
		if err := session.Delete(); err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}
	}

	// HTTP 200 OK with JSON
	ren.JSON(w, 200, ErrorResponse{})
	return
}
