package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"

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

// SearchCache stores the retrieved urls for a user
type SearchCache struct {
	sync.RWMutex
	cache map[string]map[string]*Image
}

// Store keeps a value in the searchCache
func (cache *SearchCache) Store(key string, images map[string]*Image) {
	cache.Lock()
	if cache.cache == nil {
		cache.cache = make(map[string]map[string]*Image)
	}

	if images != nil {
		cache.cache[key] = images
	}
	cache.Unlock()
}

// Retrieve gets an item from the searchCache or nil
func (cache *SearchCache) Retrieve(key string) map[string]*Image {
	var value map[string]*Image

	cache.RLock()
	if cache.cache == nil {
		value = nil
	} else {
		value = cache.cache[key]
	}
	cache.RUnlock()

	return value
}

// Environment contains all the runtime info
var Environment Settings

var reddit RedditProvider
var imgur ImgurProvider
var searchCache SearchCache

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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
		result := &Result{user, getUser(user)}
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
	if value := searchCache.Retrieve(user); value != nil {
		fmt.Println("Cache hit for", user, "with", len(value), "URLs.")
		return value
	}

	fmt.Println("Miss for", user)
	children := getChildren(user)
	fields := childrenToFields(children)
	URLs := reddit.GetURLs(fields)

	images := make(map[string]*Image)
	for img := range imgur.GetImages(URLs) {
		images[img.ID] = img
	}

	go searchCache.Store(user, images)

	return images
}
