package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

func getChildren(user string) <-chan Child {
	output := make(chan Child)

	urls := []string{
		"http://www.reddit.com/user/" + user + "/comments.json",
		"http://www.reddit.com/user/" + user + "/submitted.json",
		"http://www.reddit.com/user/" + user + ".json",
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

func childrenToFields(subs <-chan Child) <-chan string {
	out := make(chan string)
	go func() {
		for sub := range subs {
			// pull directly from post
			out <- sub.Data.URL

			// pull from comment text
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

func fieldsToURLs(input <-chan string) <-chan *url.URL {
	out := make(chan *url.URL)
	go func() {
		for value := range input {
			if imgURL, err := url.Parse(value); err == nil && imgURL.Scheme != "" {
				out <- imgURL
			}
		}
		close(out)
	}()
	return out
}
