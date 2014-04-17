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
	Songs      []data.Song   `json:"songs"`
	render     render.Render `json:"-"`
}

// RenderError renders a JSON error message with the specified HTTP status code and message
func (f *FoldersResponse) RenderError(code int, message string) {
	// Nullify all other fields
	f.Folders = nil
	f.Subfolders = nil
	f.Songs = nil

	// Generate error
	f.Error = new(Error)
	f.Error.Code = code
	f.Error.Message = message

	// Render with specified HTTP status code
	f.render.JSON(code, f)
}

// ServerError is a shortcut to render a HTTP 500 with generic "server error" message
func (f *FoldersResponse) ServerError() {
	f.RenderError(500, "server error")
	return
}

// GetFolders retrieves one or more folders from wavepipe, and returns a HTTP status and JSON
func GetFolders(r render.Render, params martini.Params) {
	// Output struct for folders request
	res := FoldersResponse{render: r}

	// List of folders to return
	folders := make([]data.Folder, 0)

	// Check API version
	if version, ok := params["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			res.RenderError(400, "unsupported API version: "+version)
			return
		}
	}

	// Check for an ID parameter
	if pID, ok := params["id"]; ok {
		// Verify valid integer ID
		id, err := strconv.Atoi(pID)
		if err != nil {
			res.RenderError(400, "invalid integer folder ID")
			return
		}

		// Load the folder
		folder := new(data.Folder)
		folder.ID = id
		if err := folder.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				res.RenderError(404, "folder ID not found")
				return
			}

			// All other errors
			log.Println(err)
			res.ServerError()
			return
		}

		// Add folder to slice
		folders = append(folders, *folder)

		// Load all subfolders
		subfolders, err := folder.Subfolders()
		if err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Add subfolders to response
		res.Subfolders = subfolders

		// Load all contained songs in this folder
		songs, err := data.DB.SongsForFolder(folder.ID)
		if err != nil {
			log.Println(err)
			res.ServerError()
			return
		}

		// Add songs to response
		res.Songs = songs
	} else {
		// Retrieve all folders
		tempFolders, err := data.DB.AllFolders()
		if err != nil {
			log.Println(err)
			res.ServerError()
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
