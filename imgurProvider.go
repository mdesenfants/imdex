package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// ImgurCache stores commonly-retrieved images
type ImgurCache struct {
	sync.RWMutex
	cache map[string][]*Image
}

var singleCache = ImgurCache{}

// ImgurProvider provides images from URLs
type ImgurProvider struct{}

type imgurImage struct {
	ID   string `json:"id"`
	Link string `json:"link"`
	NSFW bool   `json:"nsfw"`
}

type singleImage struct {
	Image imgurImage `json:"data"`
}

type imgurAlbum struct {
	Data struct {
		Images []imgurImage `json:"images"`
		Link   string       `json:"link"`
		ID     string       `json:"id"`
		NSFW   bool         `json:"nsfw"`
	} `json:"data"`
}

// GetImages produces a channel of imgur images
func (prov *ImgurProvider) GetImages(urls <-chan *url.URL) <-chan *Image {
	output := make(chan *Image)
	var wg sync.WaitGroup

	for u := range urls {
		wg.Add(1)
		go func(u *url.URL) {
			defer wg.Done()

			if strings.Contains(u.Host, "imgur.com") {
				id := getImgurID(u)
				if images := singleCache.Retrieve(id); images != nil {
					fmt.Println("Imgur cache hit for", id)
					for _, val := range images {
						output <- val
					}
					return
				}

				directory := strings.Split(u.Path, "/")[1]

				var endpoint string
				switch directory {
				case "a":
					endpoint = "album"
				case "gallery":
					endpoint = "gallery/album"
				default:
					endpoint = "image"
				}

				for val := range imgurRequest(endpoint, id) {
					output <- val
				}
			}
		}(u)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

func imgurRequest(endpoint, id string) <-chan *Image {
	images := make(chan *Image)
	if lister := singleCache.Retrieve(id); lister != nil {
		go func() {
			for _, image := range lister {
				images <- image
			}
			close(images)
		}()
		return images
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://imgur-apiv3.p.mashape.com/3/"+endpoint+"/"+id, nil)
	req.Header.Add("X-Mashape-Key", Environment.MashapeKey)
	req.Header.Add("Authorization", "Client-ID "+Environment.ImgurClientID)

	resp, err := client.Do(req)
	if err != nil {
		close(images)
		return images
	}

	if endpoint == "album" || endpoint == "gallery/album" {
		var a imgurAlbum
		dec := json.NewDecoder(resp.Body)

		if decerr := dec.Decode(&a); decerr == nil {
			go func() {
				for _, image := range a.Data.Images {
					if image.ID != "" {
						img := &Image{
							"imgur.com",
							image.ID,
							"http://i.imgur.com/" + image.ID + "m.jpg",
							a.Data.Link + "#" + image.ID,
							a.Data.NSFW,
						}
						images <- img
					}
				}
				close(images)
			}()
		} else {
			fmt.Println("Decoding error for album", decerr)
			close(images)
		}

		resp.Body.Close()
		return images
	}

	go func() {
		var si singleImage
		dec := json.NewDecoder(resp.Body)

		if decerr := dec.Decode(&si); decerr == nil && si.Image.ID != "" {
			img := &Image{
				"imgur.com",
				si.Image.ID,
				"http://i.imgur.com/" + si.Image.ID + "m.jpg",
				"http://imgur.com/" + si.Image.ID,
				si.Image.NSFW,
			}
			go singleCache.Store(img.ID, img)
			images <- img
		} else {
			fmt.Println("Decoding error for image")
			fmt.Println("\t", decerr)
			fmt.Println("\t", si)
		}

		close(images)
		resp.Body.Close()
	}()

	return images
}

func getImgurID(value *url.URL) string {
	parts := strings.Split(value.Path, "/")
	return strings.Split(parts[len(parts)-1], ".")[0]
}

// Store keeps a value in the cache
func (cache *ImgurCache) Store(key string, images ...*Image) {
	cache.Lock()
	if cache.cache == nil {
		cache.cache = make(map[string][]*Image)
	}

	if images != nil {
		if c, ok := cache.cache[key]; ok {
			c = append(c, images...)
		} else {
			cache.cache[key] = images
		}
	}
	cache.Unlock()
}

// Retrieve gets an item from the ImgurCache or nil
func (cache *ImgurCache) Retrieve(key string) []*Image {
	var value []*Image

	cache.RLock()
	if cache.cache == nil {
		value = nil
	} else {
		value = cache.cache[key]
	}
	cache.RUnlock()

	return value
}
