package main

import (
	"fmt"
	"net/url"
	"testing"
)

var directImage, _ = url.Parse("http://i.imgur.com/T37Gba0.gif")
var imagePage, _ = url.Parse("http://imgur.com/T37Gba0")
var unsupported, _ = url.Parse("http://unsupported")
var album, _ = url.Parse("http://imgur.com/a/HTxXk")

func TestGetImages(t *testing.T) {
	setup()
	urlc := make(chan *url.URL, 1)
	urlc <- imagePage
	close(urlc)

	count := 0
	for _ = range GetImages(urlc) {
		count++
	}

	if count != 1 {
		t.Error(fmt.Sprintf("Found %v image(s); expected 1.", count))
	}
}

func TestGetAlbum(t *testing.T) {
	setup()
	urlc := make(chan *url.URL, 1)
	urlc <- album
	close(urlc)

	count := 0
	for image := range GetImages(urlc) {
		fmt.Println(image)
		count++
	}

	if count != 23 {
		t.Error(fmt.Sprintf("Found %v image(s); expected 23.", count))
	}
}
