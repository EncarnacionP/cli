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

func (c *Client) GetIssues(owner, repo string, perPage int) ([]Issue, error) {
	var allIssues []Issue
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
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var issues []Issue
		if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues...)

		nextURL = getNextPageURL(resp.Header.Get("Link"))
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
