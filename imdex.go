package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
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

// LinkCache stores the retrieved urls for a user
type LinkCache struct {
	sync.RWMutex
	cache map[string][]*url.URL
}

// Store keeps a value in the linkCache
func (cache *LinkCache) Store(key string, urls ...*url.URL) {
	cache.Lock()
	if cache.cache == nil {
		cache.cache = make(map[string][]*url.URL)
	}

	if urls != nil {
		cache.cache[key] = append(cache.cache[key], urls...)
	}
	cache.Unlock()
}

// Retrieve gets an item from the linkCache or nil
func (cache *LinkCache) Retrieve(key string) []*url.URL {
	var value []*url.URL

	cache.RLock()
	if cache.cache == nil {
		value = nil
	}

	value = cache.cache[key]

	cache.RUnlock()

	return value
}

// Environment contains all the runtime info
var Environment Settings

var reddit RedditProvider
var imgur ImgurProvider
var linkCache LinkCache

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
	URLs := make(chan *url.URL)

	if values := linkCache.Retrieve(user); values != nil {
		fmt.Println("Cache hit for", user, "with", len(values), "URLs.")
		go func() {
			for _, value := range values {
				URLs <- value
			}
			close(URLs)
		}()
	} else {
		children := getChildren(user)
		fields := childrenToFields(children)

		go func() {
			for value := range cacheURLs(user, reddit.GetURLs(fields)) {
				URLs <- value
			}
			close(URLs)
		}()
	}

	images := make(map[string]*Image)
	for img := range imgur.GetImages(URLs) {
		images[img.ID] = img
	}

	return images
}

func cacheURLs(user string, u <-chan *url.URL) <-chan *url.URL {
	output := make(chan *url.URL)
	go func() {
		for val := range u {
			linkCache.Store(user, val)
			output <- val
		}

		close(output)
	}()

	return output
}
