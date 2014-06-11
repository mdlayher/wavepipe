package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/goset"
	"github.com/unrolled/render"
)

// SearchResponse repoutents the JSON outponse for /api/search
type SearchResponse struct {
	Error   *Error        `json:"error"`
	Artists []data.Artist `json:"artists"`
	Albums  []data.Album  `json:"albums"`
	Songs   []data.Song   `json:"songs"`
	Folders []data.Folder `json:"folders"`
}

// GetSearch searches for artists, albums, songs, and folders matching a specified search query,
// and returns a HTTP status and JSON
func GetSearch(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for songs request
	out := SearchResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for a search query
	query, ok := mux.Vars(req)["query"]
	if !ok {
		r.JSON(res, 400, errRes(400, "no search query specified"))
		return
	}

	// Default list of type to include in outults
	defaultTypeSet := set.New("artists", "albums", "songs", "folders")

	// Check for a comma-separated list of data types to search
	types := req.URL.Query().Get("type")
	var typeSet *set.Set
	if types == "" {
		// Search for all types if not specified
		typeSet = defaultTypeSet
	} else {
		// Iterate comma-separated list of types
		typeSet = set.New()
		for _, t := range strings.Split(types, ",") {
			// Add valid types to list
			if defaultTypeSet.Has(t) {
				typeSet.Add(t)
			}
		}
	}

	// If selected, include artists
	if typeSet.Has("artists") {
		// Search for artists which match the search query
		artists, err := data.DB.SearchArtists(query)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Copy into outponse
		out.Artists = artists
	}

	// If selected, include albums
	if typeSet.Has("albums") {
		// Search for albums which match the search query
		albums, err := data.DB.SearchAlbums(query)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Copy into outponse
		out.Albums = albums
	}

	// If selected, include songs
	if typeSet.Has("songs") {
		// Search for songs which match the search query
		songs, err := data.DB.SearchSongs(query)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Copy into outponse
		out.Songs = songs
	}

	// If selected, include folders
	if typeSet.Has("folders") {
		// Search for folders which match the search query
		folders, err := data.DB.SearchFolders(query)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Copy into outponse
		out.Folders = folders
	}

	// HTTP 200 OK with JSON
	out.Error = nil
	r.JSON(res, 200, out)
	return
}
