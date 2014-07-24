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

func getImgurId(value url.URL) string {
	parts := strings.Split(value.Path, "/")
	return strings.Replace(parts[len(parts)-1], ".jpg", "", -1)
}

func albumToImages(loc url.URL) []url.URL {
	id := getImgurId(loc)
	output := make([]url.URL, 20)

	list, err := http.Get("https://api.imgur.com/3/album/" + id)
	if err == nil {
		var a imgurAlbum
		dec := json.NewDecoder(list.Body)
		if decerr := dec.Decode(&a); decerr == nil {
			for _, image := range a.Data.Images {
				if link, err := url.Parse(image.Link); err == nil {
					output = append(output, *link)
				}
			}
		}
		list.Body.Close()
	}

	return output
}

func imgurToImage(value url.URL) Image {
	img := getImgurId(value)

	return Image{
		"imgur.com",
		img,
		fmt.Sprintf("http://i.imgur.com/%vs.jpg", img),
		fmt.Sprintf("http://imgur.com/%v", img),
	}
}
