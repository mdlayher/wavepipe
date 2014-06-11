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

// FoldersResponse represents the JSON response for /api/folders
type FoldersResponse struct {
	Error      *Error        `json:"error"`
	Folders    []data.Folder `json:"folders"`
	Subfolders []data.Folder `json:"subfolders"`
	Songs      []data.Song   `json:"songs"`
}

// GetFolders retrieves one or more folders from wavepipe, and returns a HTTP status and JSON
func GetFolders(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for folders request
	out := FoldersResponse{}

	// List of folders to return
	folders := make([]data.Folder, 0)

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := mux.Vars(req)["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			r.JSON(res, 400, errRes(400, "invalid integer folder ID"))
			return
		}

		// Load the folder
		folder := new(data.Folder)
		folder.ID = id
		if err := folder.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				r.JSON(res, 404, errRes(404, "folder ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Add folder to slice
		folders = append(folders, *folder)

		// Load all subfolders
		subfolders, err := folder.Subfolders()
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Add subfolders to response
		out.Subfolders = subfolders

		// Load all contained songs in this folder
		songs, err := data.DB.SongsForFolder(folder.ID)
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Add songs to response
		out.Songs = songs
	} else {
		// Check for a limit parameter
		if pLimit := req.URL.Query().Get("limit"); pLimit != "" {
			// Split limit into two integers
			var offset int
			var count int
			if n, err := fmt.Sscanf(pLimit, "%d,%d", &offset, &count); n < 2 || err != nil {
				r.JSON(res, 400, errRes(400, "invalid comma-separated integer pair for limit"))
				return
			}

			// Retrieve limited subset of folders
			tempFolders, err := data.DB.LimitFolders(offset, count)
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy folders into the output slice
			folders = tempFolders
		} else {
			// Retrieve all folders
			tempFolders, err := data.DB.AllFolders()
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}

			// Copy folders into the output slice
			folders = tempFolders
		}
	}

	// Build response
	out.Error = nil
	out.Folders = folders

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
