package subsonic

import (
	"encoding/xml"
	"net/http"

	"github.com/mdlayher/wavepipe/api"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

// License represents a Subsonic license
type License struct {
	XMLName xml.Name `xml:"license,omitempty"`

	Valid bool   `xml:"valid,attr"`
	Email string `xml:"email,attr"`
	Key   string `xml:"key,attr"`
	Date  string `xml:"date,attr"`
}

// GetLicense is used in Subsonic to return information about the server's license
func GetLicense(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Create a new response container with mostly blank license
	c := newContainer()
	c.License = &License{
		Valid: true,
	}

	// Write response
	r.XML(res, 200, c)
}
