package imageProviders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/mdesenfants/imdex/common"
)

// ImgurProvider provides images from URLs
type ImgurProvider struct{}

type imgurImage struct {
	ID       string `json:"id"`
	Link     string `json:"link"`
	NSFW     bool   `json:"nsfw"`
	Animated bool   `json:"animated"`
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
func (prov *ImgurProvider) GetImages(urls <-chan common.URLWithContext) <-chan *common.Image {
	output := make(chan *common.Image)
	var wg sync.WaitGroup

	for u := range urls {
		wg.Add(1)
		go func(u common.URLWithContext) {
			defer wg.Done()

			if strings.Contains(u.URL.Host, "imgur.com") {
				id := getImgurID(&u.URL)

				directory := strings.Split(u.URL.Path, "/")[1]

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
					val.Context = u.Context
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

func retrievedToImage(image imgurImage, album string, nsfw bool) *common.Image {
	var link string
	if album != "" {
		link = album + "#" + image.ID
	} else {
		link = "http://imgur.com/" + image.ID
	}

	var animated string
	if image.Animated {
		animated = "http://i.imgur.com/" + image.ID + ".gif"
	} else {
		animated = ""
	}

	img := &common.Image{
		Host:      "imgur.com",
		ID:        image.ID,
		Thumbnail: "http://i.imgur.com/" + image.ID + "m.jpg",
		URL:       link,
		NSFW:      nsfw || image.NSFW,
		Context:   "",
		Animated:  animated,
	}

	return img
}

func imgurRequest(endpoint, id string) <-chan *common.Image {
	images := make(chan *common.Image)

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://imgur-apiv3.p.mashape.com/3/"+endpoint+"/"+id, nil)
	req.Header.Add("X-Mashape-Key", common.GetMashapeKey())
	req.Header.Add("Authorization", "Client-ID "+common.GetClientID())

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
						images <- retrievedToImage(image, a.Data.Link, a.Data.NSFW)
					}
				}
				close(images)
			}()
		} else {
			fmt.Println("Decoding error for", endpoint, id)
			close(images)
		}

		resp.Body.Close()
		return images
	}

	go func() {
		var si singleImage
		dec := json.NewDecoder(resp.Body)

		if decerr := dec.Decode(&si); decerr == nil && si.Image.ID != "" {
			img := retrievedToImage(si.Image, "", si.Image.NSFW)
			images <- img
		} else {
			fmt.Println("Decoding error for image", id)
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
