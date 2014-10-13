package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"
	"github.com/mdesenfants/imdex/common"
	"github.com/mdesenfants/imdex/providers"
)

// SearchCache stores the retrieved urls for a user
type SearchCache struct {
	sync.RWMutex
	cache map[string]map[string]*common.Image
}

// Store keeps a value in the searchCache
func (cache *SearchCache) Store(key string, images map[string]*common.Image) {
	cache.Lock()
	if cache.cache == nil {
		cache.cache = make(map[string]map[string]*common.Image)
	}

	if images != nil {
		cache.cache[key] = images

		go func() {
			timer := time.NewTimer(time.Minute * 10)
			<-timer.C
			cache.Lock()
			delete(cache.cache, key)
			cache.Unlock()
			fmt.Println("Deleted", key)
		}()
	}
	cache.Unlock()
}

// Retrieve gets an item from the searchCache or nil
func (cache *SearchCache) Retrieve(key string) map[string]*common.Image {
	var value map[string]*common.Image

	cache.RLock()
	if cache.cache == nil {
		value = nil
	} else {
		value = cache.cache[key]
	}
	cache.RUnlock()

	return value
}

var reddit RedditProvider
var imgur imageProviders.ImgurProvider
var searchCache SearchCache

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// main runs the server
func main() {
	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Use(martini.Static("js"))
	m.Use(martini.Static("images"))

	m.Get("/find/stream", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		_, p, err := conn.ReadMessage()
		if err != nil {
			return
		}

		user := string(p)

		known := make(map[string]bool)

		for result := range getUserStream(user) {
			if _, exists := known[result.ID]; !exists {
				jresp, _ := json.Marshal(result)
				conn.WriteMessage(websocket.TextMessage, jresp)
				known[result.ID] = true
			}
		}

		conn.Close()
		fmt.Println("Closed connection.")
	})

	m.Get("/find/:user", func(r render.Render, p martini.Params) {
		user := p["user"]
		result := &common.Result{Name: user, Images: getUser(user)}
		r.JSON(200, *result)
	})

	m.Get("/:name", func(r render.Render, p martini.Params) {
		r.HTML(200, "index", "reddit user name")
	})

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", "reddit user name")
	})

	m.Run()
}

func getUser(user string) map[string]*common.Image {
	if value := searchCache.Retrieve(user); value != nil {
		fmt.Println("Cache hit for", user, "with", len(value), "URLs.")
		return value
	}

	fmt.Println("Miss for", user)
	children := getChildren(user)
	fields := childrenToFields(children)
	URLs := reddit.GetURLs(fields)

	images := make(map[string]*common.Image)
	for img := range imgur.GetImages(URLs) {
		images[img.ID] = img
	}

	go searchCache.Store(user, images)

	return images
}

func getUserStream(user string) <-chan *common.Image {
	if value := searchCache.Retrieve(user); value != nil {
		fmt.Println("Cache hit for", user, "with", len(value), "URLs.")

		imageChan := make(chan *common.Image)

		go func() {
			for _, img := range value {
				imageChan <- img
			}
			close(imageChan)
		}()

		return imageChan
	}

	children := getChildren(user)
	fields := childrenToFields(children)
	URLs := reddit.GetURLs(fields)

	go getUser(user)

	return imgur.GetImages(URLs)
}
