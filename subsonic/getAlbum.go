package subsonic

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

// GetAlbum is used in Subsonic to return a single album
func GetAlbum(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Fetch ID parameter
	pID := req.URL.Query().Get("id")
	if pID == "" {
		r.XML(res, 200, ErrMissingParameter)
		return
	}

	// Parse ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Load album by ID
	album := &data.Album{ID: id}
	if err := album.Load(); err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Load songs for album
	songs, err := data.DB.SongsForAlbum(album.ID)
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

	// Build and copy album container into output
	outAlbum := subAlbum(*album, songs)
	outAlbum.Songs = outSongs
	c.Album = []Album{outAlbum}

	// Write response
	r.XML(res, 200, c)
}
