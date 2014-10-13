package main

import (
	"net/url"

	"github.com/mdesenfants/imdex/common"
)

// ImageProvider turns a channel of url pointers into a channel of Image pointers
type ImageProvider interface {
	GetImages(urls <-chan *url.URL) <-chan *common.Image
}

// URLProvider provides a channel of strings into a channel of URLs
type URLProvider interface {
	GetURLs(input <-chan string) <-chan *url.URL
}
