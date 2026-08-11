package api

import (
	"net/http"
)

// FetchAllPages demonstrates the correct pagination loop logic.
func FetchAllPages(client *http.Client, initialURL string, limit int) ([]interface{}, error) {
	var allResults []interface{}
	currentURL := initialURL

	for currentURL != "" && (limit <= 0 || len(allResults) < limit) {
		req, _ := http.NewRequest("GET", currentURL, nil)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		// ... logic to decode results ...
		// results := decode(resp)
		// allResults = append(allResults, results...)

		currentURL = GetNextPageURL(resp.Header)
	}
	return allResults, nil
}
