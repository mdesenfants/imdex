package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type imgurImage struct {
	Link string `json:"link"`
}

type imgurAlbum struct {
	Data struct {
		Images []imgurImage `json:"images"`
	} `json:"data"`
}

// GetImages produces a channel of imgur images
func GetImages(urls <-chan *url.URL) <-chan *Image {
	var output <-chan *Image

	for u := range urls {
		if strings.Contains(u.Host, "imgur.com") {
			directory := strings.Split(u.Path, "/")[1]
			switch directory {
			case "a":
				output = getAlbumImages(u)
			case "gallery":
				output = getGalleryImages(u)
			default:
				single := make(chan *Image)
				go func(u *url.URL) { single <- getImage(u); close(single) }(u)
				output = single
			}
		}
	}

	return output
}

func getAlbumImages(u *url.URL) <-chan *Image {
	images := make(chan *url.URL)
	id := getImgurID(u)

	// Mashape client
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://imgur-apiv3.p.mashape.com/3/album/"+id, nil)
	req.Header.Add("X-Mashape-Key", Environment.MashapeKey)
	req.Header.Add("Authorization", "Client-ID " + Environment.ImgurClientID)

	list, err := client.Do(req)
	defer list.Body.Close()
	if err == nil {
		var a imgurAlbum
		dec := json.NewDecoder(list.Body)

		if decerr := dec.Decode(&a); decerr == nil {
			go func() {
				for _, image := range a.Data.Images {
					if link, err := url.Parse(image.Link); err == nil {
						images <- link
					}
				}
				close(images)
			}()
		}
	}

	return GetImages(images)
}

func getGalleryImages(u *url.URL) <-chan *Image {
	images := make(chan *Image)

	close(images)
	return images
}

func getImage(u *url.URL) *Image {
	imgID := getImgurID(u)

	image := &Image{
		"imgur.com",
		imgID,
		fmt.Sprintf("http://i.imgur.com/%vs.jpg", imgID),
		fmt.Sprintf("http://imgur.com/%v", imgID),
	}

	return image
}

func getImgurID(value *url.URL) string {
	parts := strings.Split(value.Path, "/")
	return strings.Replace(parts[len(parts)-1], ".jpg", "", -1)
}
