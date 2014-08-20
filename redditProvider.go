package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"unicode"
)

// RedditProvider returns image links given a context
type RedditProvider struct{}

// A Child is a reddit structure with information about a post
type Child struct {
	Kind string `json:"kind"`
	Data struct {
		ID        string `json:"id"`
		Domain    string `json:"domain"`
		URL       string `json:"url"`
		Over18    bool   `json:"over_18"`
		Body      string `json:"body"`
		Permalink string `json:"permalink"`
		LinkID    string `json:"link_id"`
	} `json:"data"`
}

// ListingData is a collection of Children
type ListingData struct {
	Children []Child `json:"children"`
}

// Listing is a reddit listing of posts
type Listing struct {
	ListingData `json:"data"`
}

// Field contains a string and the context from which it originated
type Field struct {
	Value   *string
	Context *string
}

// URLWithContext contains a url pointer and the context from which it originated
type URLWithContext struct {
	URL     *url.URL
	Context *string
}

func getChildren(user string) <-chan Child {
	output := make(chan Child)

	urls := []string{
		"http://www.reddit.com/user/" + user + "/comments.json?sort=top",
		"http://www.reddit.com/user/" + user + "/submitted.json?sort=top",
		"http://www.reddit.com/user/" + user + ".json?sort=top",
	}

	go func() {
		for _, address := range urls {
			list, err := http.Get(address)
			if err != nil {
				continue
			}

			var l Listing
			dec := json.NewDecoder(list.Body)
			if decerr := dec.Decode(&l); decerr != nil {
				close(output)
				break
			}

			for _, value := range l.Children {
				output <- value
			}

			list.Body.Close()
		}

		close(output)
	}()

	return output
}

func childrenToFields(subs <-chan Child) <-chan Field {
	out := make(chan Field)
	go func() {
		for sub := range subs {
			// pull directly from post
			var context string
			switch sub.Kind {
			case "t3":
				context = "http://reddit.com" + sub.Data.Permalink
			case "t1":
				linkID := strings.Split(sub.Data.LinkID, "_")[1]
				context = "http://www.reddit.com/comments/" + linkID + "/_/" + sub.Data.ID + "?context=3"
			}

			out <- Field{&sub.Data.URL, &context}

			// pull from comment text
			fields := strings.FieldsFunc(sub.Data.Body, func(c rune) bool {
				return unicode.IsSpace(c) || strings.ContainsRune("[]()", c)
			})

			for _, field := range fields {
				out <- Field{&field, &context}
			}
		}
		close(out)
	}()
	return out
}

// GetURLs grabs strings and parses them into urls if possible
func (red RedditProvider) GetURLs(input <-chan Field) <-chan URLWithContext {
	out := make(chan URLWithContext)
	go func() {
		var wg sync.WaitGroup
		for value := range input {
			wg.Add(1)
			go func(value Field) {
				defer wg.Done()
				if imgURL, err := url.Parse(*value.Value); err == nil && imgURL.Scheme != "" {
					out <- URLWithContext{imgURL, value.Context}
				}
			}(value)
		}
		wg.Wait()
		close(out)
	}()
	return out
}
