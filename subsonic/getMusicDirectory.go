package subsonic

import (
	"encoding/xml"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

// MusicDirectoryContainer contains a list of emulated Subsonic music folders
type MusicDirectoryContainer struct {
	// Container name
	XMLName xml.Name `xml:"directory,omitempty"`

	// Attributes
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`

	Children []Child `xml:"child"`
}

type Child struct {
	// Container name
	XMLName xml.Name `xml:"child,omitempty"`

	// Attributes
	ID       string `xml:"id,attr"`
	Title    string `xml:"title,attr"`
	Album    string `xml:"album,attr"`
	Artist   string `xml:"artist,attr"`
	IsDir    bool   `xml:"isDir,attr"`
	CoverArt int    `xml:"coverArt,attr"`
	Created  string `xml:"created,attr"`
}

// GetMusicDirectory is used in Subsonic to return a list of filesystem items
// contained in a directory, including songs, folders, etc.
func GetMusicDirectory(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Fetch ID parameter
	pID := req.URL.Query().Get("id")
	if pID == "" {
		r.XML(res, 200, ErrMissingParameter)
		return
	}

	// Parse prefix and ID in form prefix_id
	pair := strings.Split(pID, "_")
	if len(pair) < 2 {
		r.XML(res, 200, ErrMissingParameter)
		return
	}

	// Parse ID as integer
	prefix := pair[0]
	id, err := strconv.Atoi(pair[1])
	if err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Title to relay to clients
	var outTitle string

	// Create list of children to relay to clients
	children := make([]Child, 0)

	// If an artist was passed, Subsonic probably wants albums
	if prefix == "artist" {
		// Load artist to get information
		artist := &data.Artist{ID: id}
		if err := artist.Load(); err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
			return
		}

		// Use artist information for output
		outTitle = artist.Title

		// Load albums using artist ID
		albums, err := data.DB.AlbumsForArtist(id)
		if err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
			return
		}

		// Add albums to children
		for _, a := range albums {
			children = append(children, Child{
				ID:     "album_" + strconv.Itoa(a.ID),
				Title:  a.Title,
				Album:  a.Title,
				Artist: a.Artist,
				IsDir:  true,
				//CoverArt: a.ArtID,
				//Created: time.Unix(a.LastModified, 0).Format("2006-01-02T15:04:05"),
			})
		}
	}

	// If an album was passed, Subsonic probably wants songs
	if prefix == "album" {
		// Load album to get information
		album := &data.Album{ID: id}
		if err := album.Load(); err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
			return
		}

		// Use album information for output
		outTitle = album.Title

		// Load all contained songs for this album
		songs, err := data.DB.SongsForAlbum(id)
		if err != nil {
			log.Println(err)
			r.XML(res, 200, ErrGeneric)
			return
		}

		// Iterate songs and add to children
		for _, s := range songs {
			children = append(children, Child{
				ID:       strconv.Itoa(s.ID),
				Title:    s.Title,
				Album:    s.Album,
				Artist:   s.Artist,
				IsDir:    false,
				CoverArt: s.ArtID,
				Created:  time.Unix(s.LastModified, 0).Format("2006-01-02T15:04:05"),
			})
		}
	}

	// Create a new response container
	c := newContainer()
	c.MusicDirectory = &MusicDirectoryContainer{
		ID:       pID,
		Name:     outTitle,
		Children: children,
	}

	// Write response
	r.XML(res, 200, c)
}
