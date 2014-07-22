package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"net/url"
	"encoding/json"
	"strings"
	"unicode"
	"fmt"
)

type Image struct {
	Thumbnail string `json:"thumbnail"`
	Url string `json:"url"`
}

type Result struct {
	Name string 	`json:"name"`
	Images []Image	`json:"images"`
}

type Child struct {
	Data struct {
		Domain string `json:"domain"`
		Url string `json:"url"`
		Over18 bool `json:"over_18"`
		Body string `json:"body"`
	} `json:"data"`
}

type ListingData struct {
	Children []Child `json:"children"`
}

type Listing struct {
	ListingData `json:"data"`
}

var UserCache map[string]*Result = make(map[string]*Result)

func main() {
	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Use(martini.Static("js"))

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

func getGallery(user string) []Image {
	children := getChildren(user)

	images := []Image{}
	for img := range makeImages(makeUrls(getChildUrls(children))) {
		images = append(images, img)
	}

	return images
}

func getChildren(user string) []Child {
	output := make([]Child, 100)
	urls := []string{
		"http://www.reddit.com/user/"+user+"/comments.json",
		"http://www.reddit.com/user/"+user+"/submitted.json",
	}

	for _, address := range urls {
		list, err := http.Get(address)
		if err == nil {
			var l Listing
			dec := json.NewDecoder(list.Body)
			if decerr := dec.Decode(&l); decerr == nil {
				output = append(output, l.Children...)
				fmt.Println("Got", len(l.Children), "children from", address)
			}
			list.Body.Close()
		}
	}

	return output;
}

func getChildUrls(subs []Child) <-chan string {
	out := make(chan string)
	go func() {
		for _, sub := range subs {
			out<-sub.Data.Url

			fields := strings.FieldsFunc(sub.Data.Body, func(c rune) bool {
				return unicode.IsSpace(c) || strings.ContainsRune("[]()", c)
			})

			for _, field := range fields {
				out<-field
			}
		}
		close(out)
	}()
	return out
}

func makeUrls(input <-chan string) <-chan url.URL {
	out := make(chan url.URL)
	go func() {
		for value := range input {
			if imgUrl, err := url.Parse(value); err == nil && strings.Contains(imgUrl.Host, "imgur.com") {
				out<-*imgUrl
			}
		}
		close(out)
	}()
	return out
}

func makeImages(input <-chan url.URL) <-chan Image {
	out := make(chan Image)
	go func() {
		for value := range input {
			out<-toImage(value)
		}
		close(out)
	}()
	return out
}

func toImage(value url.URL) Image {
	parts := strings.Split(value.Path, "/")
	img := strings.Replace(parts[len(parts)-1], ".jpg", "", -1)

	return Image{
		fmt.Sprintf("http://i.imgur.com/%vs.jpg", img),
		fmt.Sprintf("http://imgur.com/%v", img),
	}
}
