package subsonic

import (
	"encoding/xml"
	"net/http"

	"github.com/mdlayher/wavepipe/api"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

// Starred represents a Subsonic license
type Starred struct {
	XMLName xml.Name `xml:"starred,omitempty"`
}

// GetStarred is used in Subsonic to return favorite items from the server
func GetStarred(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Create a new response container
	c := newContainer()

	// wavepipe currently has nothing like favorites/stars.  Return a blank container.
	c.Starred = &Starred{}

	// Write response
	r.XML(res, 200, c)
}
