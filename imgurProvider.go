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
	output := make(chan *Image)

	go func() {
		for u := range urls {
			if strings.Contains(u.Host, "imgur.com") {
				directory := strings.Split(u.Path, "/")[1]

				switch directory {
				case "a":
					for val := range getAlbumImages(u) {
						output <- val
					}
				case "gallery":
					for val := range getGalleryImages(u) {
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
	links := make(chan *url.URL)
	images := make(chan *Image)
	id := getImgurID(u)

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
		true,
	}

	return image
}

func getImgurID(value *url.URL) string {
	parts := strings.Split(value.Path, "/")
	return strings.Replace(parts[len(parts)-1], ".jpg", "", -1)
}
