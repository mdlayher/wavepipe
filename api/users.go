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

// UsersResponse represents the JSON response for the Users API.
type UsersResponse struct {
	Error *Error      `json:"error"`
	Users []data.User `json:"users"`
}

// GetUsers retrieves one or more users from wavepipe, and returns a HTTP status and JSON.
// It can be used to fetch a single user, or all users, depending on the request parameters.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Output struct for users request
	out := UsersResponse{}

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
			ren.JSON(w, 400, errRes(400, "invalid integer user ID"))
			return
		}

		// Load the user
		user := &data.User{ID: id}
		if err := user.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				ren.JSON(w, 404, errRes(404, "user ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// HTTP 200 OK with JSON
		out.Users = []data.User{*user}
		ren.JSON(w, 200, out)
		return
	}

	// If no other case, retrieve all users
	users, err := data.DB.AllUsers()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Users = users
	ren.JSON(w, 200, out)
	return
}

// PostUsers creates a new user for the wavepipe API, and returns a HTTP status and JSON.
func PostUsers(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Attempt to retrieve user from context
	sessionUser := new(data.User)
	if tempUser := context.Get(r, CtxUser); tempUser != nil {
		sessionUser = tempUser.(*data.User)
	} else {
		// No sessionUser stored in context
		log.Println("api: no sessionUser stored in request context!")
		ren.JSON(w, 500, serverErr)
		return
	}

	// Output struct for users request
	out := UsersResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Only allow administrators to create users
	if sessionUser.RoleID < data.RoleAdmin {
		ren.JSON(w, 403, permissionErr)
		return
	}

	// Check for required username, password, and role parameters
	username := r.PostFormValue("username")
	if username == "" {
		ren.JSON(w, 400, errRes(400, "missing required parameter: username"))
		return
	}

	password := r.PostFormValue("password")
	if password == "" {
		ren.JSON(w, 400, errRes(400, "missing required parameter: password"))
		return
	}

	// Check for role ID
	role := r.PostFormValue("role")
	if role == "" {
		ren.JSON(w, 400, errRes(400, "missing required parameter: role"))
		return
	}

	// Ensure role is valid integer, and valid role
	roleID, err := strconv.Atoi(role)
	if err != nil || (roleID != data.RoleGuest && roleID != data.RoleUser && roleID != data.RoleAdmin) {
		ren.JSON(w, 400, errRes(400, "invalid integer role ID"))
		return
	}

	// Generate a new user using the input username, password, and role
	user, err := data.NewUser(username, password, roleID)
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Users = []data.User{*user}
	ren.JSON(w, 200, out)
	return
}

// PutUsers updates users on the wavepipe API, and returns a HTTP status and JSON.
func PutUsers(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Attempt to retrieve user from context
	sessionUser := new(data.User)
	if tempUser := context.Get(r, CtxUser); tempUser != nil {
		sessionUser = tempUser.(*data.User)
	} else {
		// No sessionUser stored in context
		log.Println("api: no sessionUser stored in request context!")
		ren.JSON(w, 500, serverErr)
		return
	}

	// Output struct for users request
	out := UsersResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Disallow guests from updating any user, including themselves
	if sessionUser.RoleID == data.RoleGuest {
		ren.JSON(w, 403, permissionErr)
		return
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(r)["id"]
	if !ok {
		ren.JSON(w, 400, errRes(400, "no integer user ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		ren.JSON(w, 400, errRes(400, "invalid integer user ID"))
		return
	}

	// Load the user
	user := &data.User{ID: id}
	if err := user.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			ren.JSON(w, 404, errRes(404, "user ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Check for parameters to update the user
	if username := r.PostFormValue("username"); username != "" {
		user.Username = username
	}

	if password := r.PostFormValue("password"); password != "" {
		user.SetPassword(password)
	}

	// Check for role ID
	if role := r.PostFormValue("role"); role != "" {
		// Ensure role is valid integer, and valid role
		roleID, err := strconv.Atoi(role)
		if err != nil || (roleID != data.RoleGuest && roleID != data.RoleUser && roleID != data.RoleAdmin) {
			ren.JSON(w, 400, errRes(400, "invalid integer role ID"))
			return
		}
		user.RoleID = roleID

		// If the user is updating itself and not an administrator, do not allow a role change
		if sessionUser.RoleID != data.RoleAdmin && sessionUser.RoleID != user.RoleID {
			ren.JSON(w, 403, permissionErr)
			return
		}
	}

	// Only allow administrators to update users, unless the user is updating itself
	if sessionUser.RoleID < data.RoleAdmin && sessionUser.ID != user.ID {
		ren.JSON(w, 403, permissionErr)
		return
	}

	// Save and update the user
	if err := user.Update(); err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Users = []data.User{*user}
	ren.JSON(w, 200, out)
	return
}

// DeleteUsers deletes users from the wavepipe API, and returns a HTTP status and JSON.
func DeleteUsers(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Attempt to retrieve user from context
	user := new(data.User)
	if tempUser := context.Get(r, CtxUser); tempUser != nil {
		user = tempUser.(*data.User)
	} else {
		// No user stored in context
		log.Println("api: no user stored in request context!")
		ren.JSON(w, 500, serverErr)
		return
	}

	// Output struct for users request
	out := UsersResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Only allow administrators to delete users
	if user.RoleID < data.RoleAdmin {
		ren.JSON(w, 403, permissionErr)
		return
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(r)["id"]
	if !ok {
		ren.JSON(w, 400, errRes(400, "no integer user ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		ren.JSON(w, 400, errRes(400, "invalid integer user ID"))
		return
	}

	// Load the user
	delUser := &data.User{ID: id}
	if err := delUser.Load(); err != nil {
		// Check for invalid ID
		if err == sql.ErrNoRows {
			ren.JSON(w, 404, errRes(404, "user ID not found"))
			return
		}

		// All other errors
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Verify user is not attempting to delete itself
	if user.ID == delUser.ID {
		ren.JSON(w, 403, errRes(403, "cannot delete current user"))
		return
	}

	// Fetch all sessions for the user
	sessions, err := data.DB.SessionsForUser(delUser.ID)
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Iterate and delete all sessions
	for _, s := range sessions {
		// Delete the session
		if err := s.Delete(); err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}
	}

	// Delete the user
	if err := delUser.Delete(); err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// HTTP 200 OK with JSON
	out.Users = []data.User{*user}
	ren.JSON(w, 200, out)
	return
}
