package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// FoldersResponse represents the JSON response for the Folders API.
type FoldersResponse struct {
	Error      *Error        `json:"error"`
	Folders    []data.Folder `json:"folders"`
	Subfolders []data.Folder `json:"subfolders"`
	Songs      []data.Song   `json:"songs"`
}

// GetFolders retrieves one or more folders from wavepipe, and returns a HTTP status and JSON.
// It can be used to fetch a single folder, a limited subset of folders, or all folders, depending
// on the request parameters.
func GetFolders(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Output struct for folders request
	out := FoldersResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := mux.Vars(r)["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			ren.JSON(w, 400, errRes(400, "invalid integer folder ID"))
			return
		}

		// Load the folder
		folder := &data.Folder{ID: id}
		if err := folder.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				ren.JSON(w, 404, errRes(404, "folder ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Add folder to output
		out.Folders = []data.Folder{*folder}

		// Load all subfolders
		subfolders, err := folder.Subfolders()
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Add subfolders to response
		out.Subfolders = subfolders

		// Load all contained songs in this folder
		songs, err := data.DB.SongsForFolder(folder.ID)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Add songs to response
		out.Songs = songs

		// HTTP 200 OK with JSON
		ren.JSON(w, 200, out)
		return
	}

	// Check for a limit parameter
	if pLimit := r.URL.Query().Get("limit"); pLimit != "" {
		// Split limit into two integers
		var offset int
		var count int
		if n, err := fmt.Sscanf(pLimit, "%d,%d", &offset, &count); n < 2 || err != nil {
			ren.JSON(w, 400, errRes(400, "invalid comma-separated integer pair for limit"))
			return
		}

		// Retrieve limited subset of folders
		folders, err := data.DB.LimitFolders(offset, count)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Folders = folders
		ren.JSON(w, 200, out)
		return
	}

	// If no other case, retrieve all folders
	folders, err := data.DB.AllFolders()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Folders = folders
	ren.JSON(w, 200, out)
	return
}
