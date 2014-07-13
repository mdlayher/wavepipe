package subsonic

import (
	"encoding/xml"
	"net/http"

	"github.com/mdlayher/wavepipe/api"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

// Playlists represents the Subsonic playlists container
type Playlists struct {
	XMLName xml.Name `xml:"playlists,omitempty"`
}

// GetPlaylists is used in Subsonic to return playlists from the server
func GetPlaylists(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Create a new response container
	c := newContainer()

	// wavepipe currently has no playlists.  Return a blank container.
	c.Playlists = &Playlists{}

	// Write response
	r.XML(res, 200, c)
}
