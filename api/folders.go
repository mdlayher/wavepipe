package api

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// FoldersResponse represents the JSON response for /api/folders
type FoldersResponse struct {
	Error      *Error        `json:"error"`
	Folders    []data.Folder `json:"folders"`
	Subfolders []data.Folder `json:"subfolders"`
}

// GetFolders retrieves one or more folders from wavepipe, and returns a HTTP status and JSON
func GetFolders(r render.Render, params martini.Params) {
	// Output struct for folders request
	res := FoldersResponse{}

	// List of folders to return
	folders := make([]data.Folder, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "unsupported API version: " + version
			r.JSON(400, res)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.Error = new(Error)
			res.Error.Code = 400
			res.Error.Message = "invalid integer folder ID"
			r.JSON(400, res)
			return
		}

		// Load the folder
		folder := new(data.Folder)
		folder.ID = id
		if err := folder.Load(); err != nil {
			res.Error = new(Error)

			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.Error.Code = 404
				res.Error.Message = "folder ID not found"
				r.JSON(404, res)
				return
			}

			// All other errors
			log.Println(err)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Add folder to slice
		folders = append(folders, *folder)

		// Load all subfolders
		subfolders, err := folder.Subfolders()
		if err != nil {
			log.Println(err)

			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Add subfolders to response
		res.Subfolders = subfolders
	} else {
		// Retrieve all folders
		tempFolders, err := data.DB.AllFolders()
		if err != nil {
			log.Println(err)
			res.Error = new(Error)
			res.Error.Code = 500
			res.Error.Message = "server error"
			r.JSON(500, res)
			return
		}

		// Copy folders into the output slice
		folders = tempFolders
	}

	// Build response
	res.Error = nil
	res.Folders = folders

	// HTTP 200 OK with JSON
	r.JSON(200, res)
	return
}
