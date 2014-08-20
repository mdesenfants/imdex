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
var gallery, _ = url.Parse("http://imgur.com/gallery/acXV0")
var user = "awildsketchappeared"

func TestGetImages(t *testing.T) {
	setup()
	urlc := make(chan *url.URL, 1)
	urlc <- imagePage
	close(urlc)

	count := 0
	for _ = range imgur.GetImages(urlc) {
		count++
	}

	if count != 1 {
		t.Error(fmt.Sprintf("Found %v image(s); expected 1.", count))
	}
}

func TestGetAlbum(t *testing.T) {
	setup()

	// album
	urlc := make(chan *url.URL, 1)
	urlc <- album
	close(urlc)

	count := 0
	for _ = range imgur.GetImages(urlc) {
		count++
	}

	if count != 23 {
		t.Error(fmt.Sprintf("Found %v image(s); expected 23.", count))
	}

	// gallery
	urlc = make(chan *url.URL, 1)
	urlc <- gallery
	close(urlc)

	count = 0
	for _ = range imgur.GetImages(urlc) {
		count++
	}

	if count != 36 {
		t.Error(fmt.Sprintf("Found %v image(s); expected 36.", count))
	}
}

func TestGetChildren(t *testing.T) {
	results := getChildren(user)

	count := 0
	for _ = range results {
		count++
	}

	if count < 5 {
		t.Error(fmt.Sprintf("Expected at least 5 links; found %v", count))
	}
}

func TestChildrenToFields(t *testing.T) {
	children := getChildren(user)
	fields := childrenToFields(children)

	count := 0
	for _ = range fields {
		count++
	}

	if count < 20 {
		t.Error(fmt.Sprintf("Expected a lot of fields, got %v", count))
	}
}

func TestFieldsToURLs(t *testing.T) {
	children := getChildren(user)
	fields := childrenToFields(children)
	urls := reddit.GetURLs(fields)

	count := 0
	for _ = range urls {
		count++
	}

	if count < 10 {
		t.Error(fmt.Sprintf("Expected at least 10 urls, got %v", count))
	}
}

func TestGetUser(t *testing.T) {
	// Test cache miss
	results := getUser(user)

	count := len(results)
	if count < 10 {
		t.Error(fmt.Sprintf("Expected at least 10 images, got %v", count))
	}
}

func TestGetUserStream(t *testing.T) {
	// Test cache miss
	results := getUserStream(user)

	count := 0

	for _ = range results {
		count++
	}

	if count < 10 {
		t.Error(fmt.Sprintf("Expected at least 10 images, got %v", count))
	}
}
