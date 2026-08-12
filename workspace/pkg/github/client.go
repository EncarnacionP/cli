package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type Issue struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	}
}

func (c *Client) GetIssues(owner, repo string, limit int) ([]Issue, error) {
	var allIssues []Issue

	// Per-page size. A non-positive limit means "no limit"; use a sane page
	// size in that case. GitHub caps per_page at 100, so clamp to avoid
	// unnecessary round trips (a limit of 5000 would otherwise still fetch
	// 100/page but iterate far more pages than needed).
	perPage := limit
	if perPage <= 0 {
		perPage = 100
	}
	if perPage > 100 {
		perPage = 100
	}

	nextURL := fmt.Sprintf("%s/repos/%s/%s/issues?per_page=%d", c.BaseURL, owner, repo, perPage)

	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var issues []Issue
		err = json.NewDecoder(resp.Body).Decode(&issues)
		nextLink := resp.Header.Get("Link")
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues...)

		// Stop when the user's limit is reached.
		if limit > 0 && len(allIssues) >= limit {
			allIssues = allIssues[:limit]
			break
		}

		// Safety fallback: an empty page must stop pagination even if a next
		// link is present, preventing an infinite loop against a misbehaving
		// API that keeps returning empty pages with a next relation.
		if len(issues) == 0 {
			break
		}

		// Continue strictly based on the Link header's rel="next" relation,
		// never on the number of items returned (a sparse page can carry a
		// next link with fewer items than requested).
		nextURL = getNextPageURL(nextLink)
	}

	return allIssues, nil
}

func getNextPageURL(linkHeader string) string {
	re := regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)
	matches := re.FindStringSubmatch(linkHeader)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
