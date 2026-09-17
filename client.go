package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Item represents an item returned from the API.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PageResponse represents a paginated API response body.
type PageResponse struct {
	Items []Item `json:"items"`
}

// Client handles API communication.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Client instance.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	}
}

// FetchAllItems retrieves all items across all pages.
// Fixed premature termination when server returns fewer items per page than requestedPageSize.
func (c *Client) FetchAllItems(ctx context.Context, requestedPageSize int) ([]Item, error) {
	var allItems []Item
	page := 1

	for {
		items, err := c.FetchPage(ctx, page, requestedPageSize)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		allItems = append(allItems, items...)
		page++
	}

	return allItems, nil
}

// FetchPage retrieves a single page of items from the API.
func (c *Client) FetchPage(ctx context.Context, page, pageSize int) ([]Item, error) {
	u, err := url.Parse(fmt.Sprintf("%s/items", c.BaseURL))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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

	var pageResp PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, err
	}

	return pageResp.Items, nil
}
