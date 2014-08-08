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
	Link string `json:"link"`
}

type imgurAlbum struct {
	Data struct {
		Images []imgurImage `json:"images"`
	} `json:"data"`
}

// GetImages produces a channel of imgur images
func (prov *ImgurProvider) GetImages(urls <-chan *url.URL) <-chan *Image {
	output := make(chan *Image)

	go func() {
		for u := range urls {
			if strings.Contains(u.Host, "imgur.com") {

				id := getImgurID(u)
				if images := singleCache.Retrieve(id); images != nil {
					for _, val := range images {
						output <- val
					}
					continue
				}

				directory := strings.Split(u.Path, "/")[1]

				switch directory {
				case "a":
					for val := range getAlbumImages(u) {
						output <- val
					}
				case "gallery":
					for val := range getAlbumImages(u) {
						output <- val
					}
				default:
					output <- getImage(u)
				}
			}
		}
		close(output)
	}()

	return output
}

func getAlbumImages(u *url.URL) <-chan *Image {
	id := getImgurID(u)
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

	links := make(chan *url.URL)

	// Mashape client
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://imgur-apiv3.p.mashape.com/3/album/"+id, nil)
	req.Header.Add("X-Mashape-Key", Environment.MashapeKey)
	req.Header.Add("Authorization", "Client-ID "+Environment.ImgurClientID)

	list, err := client.Do(req)
	defer list.Body.Close()
	if err != nil {
		close(links)
		close(images)
		return images
	}

	var a imgurAlbum
	dec := json.NewDecoder(list.Body)

	if decerr := dec.Decode(&a); decerr == nil {
		go func() {
			for _, image := range a.Data.Images {
				if link, err := url.Parse(image.Link); err == nil {
					links <- link
				}
			}
			close(links)
		}()
	}

	go func() {
		for l := range links {
			images <- getImage(l)
		}
		close(images)
	}()

	return images
}

func getImage(u *url.URL) *Image {
	imgID := getImgurID(u)

	image := &Image{
		"imgur.com",
		imgID,
		fmt.Sprintf("http://i.imgur.com/%vm.jpg", imgID),
		fmt.Sprintf("http://imgur.com/%v", imgID),
		true,
	}

	singleCache.Store(imgID, image)

	return image
}

func getImgurID(value *url.URL) string {
	parts := strings.Split(value.Path, "/")
	return strings.Replace(parts[len(parts)-1], ".jpg", "", -1)
}

// Store keeps a value in the cache
func (cache *ImgurCache) Store(key string, images ...*Image) {
	cache.Lock()
	if cache.cache == nil {
		cache.cache = make(map[string][]*Image)
	}

	if images != nil {
		cache.cache[key] = images
	}
	cache.Unlock()
}

// Retrieve gets an item from the ImgurCache or nil
func (cache *ImgurCache) Retrieve(key string) []*Image {
	cache.Lock()
	if cache.cache == nil {
		cache.cache = make(map[string][]*Image)
	}
	cache.Unlock()

	cache.RLock()
	value := cache.cache[key]
	cache.RUnlock()

	return value
}
