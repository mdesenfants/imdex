package common

// Image contains information about an image result
type Image struct {
	Host      string `json:"host"`
	ID        string `json:"id"`
	Thumbnail string `json:"thumbnail"`
	URL       string `json:"url"`
	NSFW      bool   `json:"nsfw"`
	Context   string `json:"context"`
	Animated  string `json:"animated"`
}
