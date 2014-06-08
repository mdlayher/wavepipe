package subsonic

import (
	"encoding/xml"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/martini-contrib/render"
)

// AlbumList2Container contains a list of emulated Subsonic albums, by tags
type AlbumList2Container struct {
	// Container name
	XMLName xml.Name `xml:"albumList2,omitempty"`

	// Albums
	Albums []Album `xml:"album"`
}

// GetAlbumList2 is used in Subsonic to return a list of albums organized with tags
func GetAlbumList2(req *http.Request, res http.ResponseWriter, r render.Render) {
	// Create a new response container
	c := newContainer()

	// Attempt to parse offset if applicable
	var offset int
	if qOffset := req.URL.Query().Get("offset"); qOffset != "" {
		// Parse integer
		tempOffset, err := strconv.Atoi(qOffset)
		if err != nil {
			log.Println(err)
			r.XML(200, ErrGeneric)
			return
		}

		// Store for use
		offset = tempOffset
	}

	// Attempt to parse size if applicable
	var size = 10
	if qSize := req.URL.Query().Get("size"); qSize != "" {
		// Parse integer
		tempSize, err := strconv.Atoi(qSize)
		if err != nil {
			log.Println(err)
			r.XML(200, ErrGeneric)
			return
		}

		// Store for use
		size = tempSize
	}

	// Fetch slice of albums to convert to Subsonic form
	albums, err := data.DB.LimitAlbums(offset, size)
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Iterate all albums
	outAlbums := make([]Album, 0)
	for _, a := range albums {
		// Load songs for album
		songs, err := data.DB.SongsForAlbum(a.ID)
		if err != nil {
			log.Println(err)
			r.XML(200, ErrGeneric)
			return
		}

		// If no songs, skip output
		if len(songs) == 0 {
			continue
		}

		// Append Subsonic album
		outAlbums = append(outAlbums, subAlbum(a, songs))
	}

	// Copy albums list into output
	c.AlbumList2 = &AlbumList2Container{Albums: outAlbums}

	// Write response
	r.XML(200, c)
}
