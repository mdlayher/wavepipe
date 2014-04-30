package data

import (
	"encoding/json"

	"code.google.com/p/go.crypto/bcrypt"
)

// User represents an user registered to wavepipe
type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	LastFMToken string `db:"lastfm_token" json:"lastfmToken"`
}

// NewUser generates and saves a new user, while also hashing the input password
func NewUser(username string, password string) (*User, error) {
	// Generate user
	user := new(User)
	user.Username = username

	// Hash input password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		return nil, err
	}
	user.Password = string(hash)

	// Save user
	if err := user.Save(); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateSession generates a new API session for this user
func (u User) CreateSession(client string) (*Session, error) {
	return NewSession(u.ID, u.Password, client)
}

// Delete removes an existing User from the database
func (u *User) Delete() error {
	return DB.DeleteUser(u)
}

// Load pulls an existing User from the database
func (u *User) Load() error {
	return DB.LoadUser(u)
}

// Save creates a new User in the database
func (u *User) Save() error {
	return DB.SaveUser(u)
}

// Update updates an existing User in the database
func (u *User) Update() error {
	return DB.UpdateUser(u)
}

// ToJSON generates a JSON representation of a User
func (u User) ToJSON() ([]byte, error) {
	// Marshal into JSON
	out, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	// Return JSON
	return out, nil
}

// FromJSON generates a User from its JSON representation
func (u *User) FromJSON(in []byte) error {
	return json.Unmarshal(in, &u)
}
