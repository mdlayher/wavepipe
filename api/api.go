package api

// Error represents an error produced by the API
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
