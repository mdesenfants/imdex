package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// Image contains information about an image result
type Image struct {
	Host      string `json:"host"`
	ID        string `json:"id"`
	Thumbnail string `json:"thumbnail"`
	URL       string `json:"url"`
}

// A Result is a list of images for a user
type Result struct {
	Name   string            `json:"name"`
	Images map[string]*Image `json:"images"`
}

// A Child is a reddit structure with information about a post
type Child struct {
	Data struct {
		Domain string `json:"domain"`
		URL    string `json:"url"`
		Over18 bool   `json:"over_18"`
		Body   string `json:"body"`
	} `json:"data"`
}

// ListingData is a collection of Children
type ListingData struct {
	Children []Child `json:"children"`
}

// Listing is a reddit listing of posts
type Listing struct {
	ListingData `json:"data"`
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
			result = &Result{user, getGallery(user)}
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

func getGallery(user string) map[string]*Image {
	children := getChildren(user)
	fields := childrenToFields(children)
	URLs := fieldsToURLs(fields)

	images := make(map[string]*Image)
	for img := range GetImages(URLs) {
		images[img.ID] = img
	}

	return images
}

func getChildren(user string) []Child {
	output := make([]Child, 100)
	urls := []string{
		"http://www.reddit.com/user/" + user + "/comments.json",
		"http://www.reddit.com/user/" + user + "/submitted.json",
		"http://www.reddit.com/user/" + user + ".json",
	}

	for _, address := range urls {
		list, err := http.Get(address)
		if err == nil {
			var l Listing
			dec := json.NewDecoder(list.Body)
			if decerr := dec.Decode(&l); decerr == nil {
				output = append(output, l.Children...)
			}
			list.Body.Close()
		}
	}

	return output
}

func childrenToFields(subs []Child) <-chan string {
	out := make(chan string)
	go func() {
		for _, sub := range subs {
			out <- sub.Data.URL

			fields := strings.FieldsFunc(sub.Data.Body, func(c rune) bool {
				return unicode.IsSpace(c) || strings.ContainsRune("[]()", c)
			})

			for _, field := range fields {
				out <- field
			}
		}
		close(out)
	}()
	return out
}

func fieldsToURLs(input <-chan string) <-chan *url.URL {
	out := make(chan *url.URL)
	go func() {
		for value := range input {
			if imgURL, err := url.Parse(value); err == nil && imgURL.Scheme != "" {
				out <- imgURL
			}
		}
		close(out)
	}()
	return out
}
