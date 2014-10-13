package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Settings contains the web app settings
type settings struct {
	ImgurClientID string `json:"imgurClientId"`
	MashapeKey    string `json:"mashapeKey"`
}

var environment *settings
var lastModified time.Time

//
func refreshSettings() {
	filename := "imdex.json"
	stat, serr := os.Stat(filename)

	var updated time.Time
	if serr != nil {
		updated = stat.ModTime()
	}

	if environment == nil || updated.After(lastModified) {
		// Setup runtime settings
		file, ferr := os.Open(filename)
		defer file.Close()
		if ferr != nil {
			panic("Could not load imdex.json")
		}

		envDec := json.NewDecoder(file)
		if decerr := envDec.Decode(&environment); decerr != nil {
			panic(fmt.Sprintf("Could not read imdex.json: %v", decerr))
		}
		lastModified = updated
	}
}

// GetClientID gets the client id for imgur
func GetClientID() string {
	refreshSettings()
	return environment.ImgurClientID
}

// GetMashapeKey gets the api key for mashape (imgur api)
func GetMashapeKey() string {
	refreshSettings()
	return environment.MashapeKey
}
