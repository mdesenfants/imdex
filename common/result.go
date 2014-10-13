package common

// A Result is a list of images for a user
type Result struct {
	Name   string            `json:"name"`
	Images map[string]*Image `json:"images"`
}
