package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// Image contains information about an image result
type Image struct {
	Host      string `json:"host"`
	ID        string `json:"id"`
	Thumbnail string `json:"thumbnail"`
	URL       string `json:"url"`
	NSFW      bool   `json:"nsfw"`
}

// A Result is a list of images for a user
type Result struct {
	Name   string            `json:"name"`
	Images map[string]*Image `json:"images"`
}

// Settings contains the web app settings
type Settings struct {
	ImgurClientID string `json:"imgurClientId"`
	MashapeKey    string `json:"mashapeKey"`
}

// UserCache contains all the information gathered about requests so far
var UserCache = make(map[string]*Result)

// Environment contains all the runtime info
var Environment Settings

var reddit RedditProvider
var imgur ImgurProvider

// main runs the server
func main() {
	m := martini.Classic()

	setup()

	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Use(martini.Static("js"))
	m.Use(martini.Static("images"))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", "reddit user name")
	})

	m.Get("/find/:user", func(r render.Render, p martini.Params) {
		user := p["user"]

		var result *Result
		var ok bool

		if result, ok = UserCache[user]; !ok {
			result = &Result{user, getUser(user)}
			UserCache[user] = result
		}

		r.JSON(200, *result)
	})

	m.Run()
}

func setup() {
	// Setup runtime settings
	file, ferr := os.Open("imdex.json")
	defer file.Close()
	if ferr != nil {
		panic("Could not load imdex.json")
	}

	envDec := json.NewDecoder(file)
	if decerr := envDec.Decode(&Environment); decerr != nil {
		panic(fmt.Sprintf("Could not read imdex.json: %v", decerr))
	}
}

func getUser(user string) map[string]*Image {
	children := getChildren(user)
	fields := childrenToFields(children)
	URLs := reddit.GetURLs(fields)

	images := make(map[string]*Image)
	for img := range imgur.GetImages(URLs) {
		images[img.ID] = img
	}

	return images
}
