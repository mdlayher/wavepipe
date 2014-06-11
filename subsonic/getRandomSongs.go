package subsonic

import (
	"encoding/xml"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

// RandomSongsContainer contains a random list of emulated Subsonic songs
type RandomSongsContainer struct {
	// Container name
	XMLName xml.Name `xml:"randomSongs,omitempty"`

	// Songs
	Songs []Song `xml:"song"`
}

// GetRandomSongs is used in Subsonic to return a list of random songs
func GetRandomSongs(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Fetch size parameter if passed
	size := 10
	if pSize := req.URL.Query().Get("size"); pSize != "" {
		// Parse size
		tempSize, err := strconv.Atoi(pSize)
		if err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
			return
		}

		size = tempSize
	}

	// Load specified size of random songs
	songs, err := data.DB.RandomSongs(size)
	if err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Create slice of Subsonic songs
	outSongs := make([]Song, 0)
	for _, s := range songs {
		outSongs = append(outSongs, subSong(s))
	}

	// Create a new response container
	c := newContainer()
	c.RandomSongs = &RandomSongsContainer{Songs: outSongs}

	// Write response
	r.XML(res, 200, c)
}
