package main

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

type Image struct {
	Host      string `json:"host"`
	Id        string `json:"id"`
	Thumbnail string `json:"thumbnail"`
	Url       string `json:"url"`
}

type Result struct {
	Name   string           `json:"name"`
	Images map[string]Image `json:"images"`
}

type Child struct {
	Data struct {
		Domain string `json:"domain"`
		Url    string `json:"url"`
		Over18 bool   `json:"over_18"`
		Body   string `json:"body"`
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
	m.Use(martini.Static("images"))

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

func getGallery(user string) map[string]Image {
	children := getChildren(user)

	images := make(map[string]Image)
	for img := range makeImages(fieldsToUrls(childrenToFields(children))) {
		images[img.Id] = img
	}

	return images
}

func getChildren(user string) []Child {
	output := make([]Child, 100)
	urls := []string{
		"http://www.reddit.com/user/" + user + "/comments.json",
		"http://www.reddit.com/user/" + user + "/submitted.json",
		"http://www.reddit.com/user/" + user + ".json",
	}

	for _, address := range urls {
		list, err := http.Get(address)
		if err == nil {
			var l Listing
			dec := json.NewDecoder(list.Body)
			if decerr := dec.Decode(&l); decerr == nil {
				output = append(output, l.Children...)
			}
			list.Body.Close()
		}
	}

	return output
}

func childrenToFields(subs []Child) <-chan string {
	out := make(chan string)
	go func() {
		for _, sub := range subs {
			out <- sub.Data.Url

			fields := strings.FieldsFunc(sub.Data.Body, func(c rune) bool {
				return unicode.IsSpace(c) || strings.ContainsRune("[]()", c)
			})

			for _, field := range fields {
				out <- field
			}
		}
		close(out)
	}()
	return out
}

func fieldsToUrls(input <-chan string) <-chan url.URL {
	out := make(chan url.URL)
	go func() {
		for value := range input {
			if imgUrl, err := url.Parse(value); err == nil && imgUrl.Scheme != "" {
				out <- *imgUrl
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
			if strings.Contains(value.Path, "/a/") {
				for _, img := range albumToImages(value) {
					out <- imgurToImage(img)
				}
			}
			out <- imgurToImage(value)
		}
		close(out)
	}()
	return out
}
