package api

import (
	"net/http"
	"regexp"
)

var nextLinkRegex = regexp.MustCompile(`<(.*?)>; rel="next"`)

// GetNextPageURL extracts the next page URL from the Link header.
func GetNextPageURL(header http.Header) string {
	link := header.Get("Link")
	if link == "" {
		return ""
	}
	matches := nextLinkRegex.FindStringSubmatch(link)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
