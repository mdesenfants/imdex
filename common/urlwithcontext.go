package common

import "net/url"

// URLWithContext contains a url pointer and the context from which it originated
type URLWithContext struct {
	URL     url.URL
	Context string
}
