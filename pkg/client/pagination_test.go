package client_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/EncarnacionP/cli/pkg/client"
)

type Item struct {
	ID   int
	Name string
}

func TestFetchAllPages_SparseAndEmptyPages(t *testing.T) {
	// Scenario:
	// Page 1 (per_page=5): returns 2 items (< per_page!), but has rel="next" Link header pointing to page 2.
	// Page 2 (per_page=5): returns 0 items, but has rel="next" Link header pointing to page 3.
	// Page 3 (per_page=5): returns 3 items, no rel="next" Link header (final page).
	// Total items returned across all pages should be 2 + 0 + 3 = 5 items over 3 requests.

	requestsMade := 0

	fetcher := func(ctx context.Context, page int, perPage int) ([]Item, *http.Response, error) {
		requestsMade++
		resp := &http.Response{Header: make(http.Header)}

		switch page {
		case 1:
			resp.Header.Set("Link", `<https://api.github.com/resource?page=2&per_page=5>; rel="next", <https://api.github.com/resource?page=3&per_page=5>; rel="last"`)
			return []Item{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}}, resp, nil
		case 2:
			resp.Header.Set("Link", `<https://api.github.com/resource?page=3&per_page=5>; rel="next"`)
			return []Item{}, resp, nil
		case 3:
			// Final page, no "next" link
			resp.Header.Set("Link", `<https://api.github.com/resource?page=1&per_page=5>; rel="first"`)
			return []Item{{ID: 3, Name: "Item 3"}, {ID: 4, Name: "Item 4"}, {ID: 5, Name: "Item 5"}}, resp, nil
		default:
			return nil, resp, fmt.Errorf("unexpected page %d", page)
		}
	}

	items, err := client.FetchAllPages(context.Background(), fetcher, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestsMade != 3 {
		t.Errorf("expected 3 requests, got %d", requestsMade)
	}

	if len(items) != 5 {
		t.Errorf("expected 5 items accumulated, got %d", len(items))
	}
}

func TestFetchAllPages_MaxItemsLimit(t *testing.T) {
	fetcher := func(ctx context.Context, page int, perPage int) ([]Item, *http.Response, error) {
		resp := &http.Response{Header: make(http.Header)}
		resp.Header.Set("Link", fmt.Sprintf(`<https://api.github.com/resource?page=%d&per_page=%d>; rel="next"`, page+1, perPage))
		return []Item{{ID: page*10 + 1}, {ID: page*10 + 2}}, resp, nil
	}

	items, err := client.FetchAllPages(context.Background(), fetcher, 2, client.PaginationOptions{MaxItems: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("expected 3 items capped by MaxItems, got %d", len(items))
	}
}

func TestFetchAllPagesGraphQL_SparsePage(t *testing.T) {
	requestsMade := 0

	fetcher := func(ctx context.Context, cursor string, perPage int) ([]Item, client.GraphQLPageInfo, error) {
		requestsMade++
		switch cursor {
		case "":
			return []Item{{ID: 101}}, client.GraphQLPageInfo{HasNextPage: true, EndCursor: "cursor_1"}, nil
		case "cursor_1":
			return []Item{{ID: 102}, {ID: 103}}, client.GraphQLPageInfo{HasNextPage: false, EndCursor: "cursor_2"}, nil
		default:
			return nil, client.GraphQLPageInfo{}, fmt.Errorf("unexpected cursor %s", cursor)
		}
	}

	items, err := client.FetchAllPagesGraphQL(context.Background(), fetcher, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestsMade != 2 {
		t.Errorf("expected 2 requests, got %d", requestsMade)
	}

	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}
