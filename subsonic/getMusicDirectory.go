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

	// Check for right prefix
	if pair[0] != "folder" {
		r.XML(res, 200, ErrMissingParameter)
		return
	}

	// Parse ID as integer
	id, err := strconv.Atoi(pair[1])
	if err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Load folder by ID
	folder := &data.Folder{ID: id}
	if err := folder.Load(); err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Load all subfolders
	subfolders, err := folder.Subfolders()
	if err != nil {
		log.Println(err)
		r.XML(res, 200, ErrGeneric)
		return
	}

	// Begin building list of children
	children := make([]Child, 0)
	for _, sf := range subfolders {
		children = append(children, Child{
			ID:    "folder_" + strconv.Itoa(sf.ID),
			Title: sf.Title,
			IsDir: true,
		})
	}

	// Load all contained songs in this folder
	songs, err := data.DB.SongsForFolder(folder.ID)
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

	/*
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
		//
	*/

	// Create a new response container
	c := newContainer()
	c.MusicDirectory = &MusicDirectoryContainer{
		ID:       "folder_" + strconv.Itoa(folder.ID),
		Name:     folder.Title,
		Children: children,
	}

	// Write response
	r.XML(res, 200, c)
}
