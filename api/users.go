package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// UsersResponse represents the JSON response for /api/users
type UsersResponse struct {
	Error *Error      `json:"error"`
	Users []data.User `json:"users"`
}

// GetUsers retrieves one or more users from wavepipe, and returns a HTTP status and JSON
func GetUsers(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for users request
	out := UsersResponse{}

	// List of users to return
	users := make([]data.User, 0)

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
			r.JSON(res, 400, errRes(400, "invalid integer user ID"))
			return
		}

		// Load the user
		user := new(data.User)
		user.ID = id
		if err := user.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				r.JSON(res, 404, errRes(404, "user ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Add user to slice
		users = append(users, *user)
	} else {
		// Retrieve all users
		tempUsers, err := data.DB.AllUsers()
		if err != nil {
			log.Println(err)
			r.JSON(res, 500, serverErr)
			return
		}

		// Copy users into the output slice
		users = tempUsers
	}

	// Build response
	out.Error = nil
	out.Users = users

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}

// PostUsers creates a new user on the wavepipe API, and returns a HTTP status and JSON
func PostUsers(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for users request
	out := UsersResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for required username and password parameters
	username := req.PostFormValue("username")
	if username == "" {
		r.JSON(res, 400, errRes(400, "missing required parameter: username"))
		return
	}

	password := req.PostFormValue("password")
	if password == "" {
		r.JSON(res, 400, errRes(400, "missing required parameter: password"))
		return
	}

	// Generate a new user using the input username and password
	user, err := data.NewUser(username, password)
	if err != nil {
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Build response
	out.Users = append(out.Users, *user)
	out.Error = nil

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}

// DeleteUsers deletes users from wavepipe, and returns a HTTP status and JSON
func DeleteUsers(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Attempt to retrieve user from context
	user := new(data.User)
	if tempUser := context.Get(req, CtxUser); tempUser != nil {
		user = tempUser.(*data.User)
	} else {
		// No user stored in context
		log.Println("api: no user stored in request context!")
		r.JSON(res, 500, serverErr)
		return
	}

	// Output struct for users request
	out := UsersResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(req)["id"]
	if !ok {
		r.JSON(res, 400, errRes(400, "no integer user ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		r.JSON(res, 400, errRes(400, "invalid integer user ID"))
		return
	}

	// Load the user
	delUser := new(data.User)
	delUser.ID = id
	if err := delUser.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			r.JSON(res, 404, errRes(404, "user ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Verify user is not attempting to delete itself
	if user.ID == delUser.ID {
		r.JSON(res, 403, errRes(403, "cannot delete current user"))
		return
	}

	// Delete the user
	if err := delUser.Delete(); err != nil {
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Build response
	out.Error = nil
	out.Users = append(out.Users, *delUser)

	// HTTP 200 OK with JSON
	r.JSON(res, 200, out)
	return
}
