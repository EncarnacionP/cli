package client

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

var linkNextRegex = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// ItemFetcher defines a function that fetches a single page of items given a page number and page size.
// It returns the list of items for that page, the HTTP response, and any error.
type ItemFetcher[T any] func(ctx context.Context, page int, perPage int) ([]T, *http.Response, error)

// PaginationOptions controls the behavior of the pagination loop.
type PaginationOptions struct {
	MaxItems int // Optional limit on total items to accumulate (0 or negative means unlimited)
}

// FetchAllPages iterates through paginated endpoints according to HTTP Link header rel="next".
// It continues fetching as long as rel="next" is present in the Link header (or until MaxItems is reached),
// regardless of whether individual pages return fewer items than perPage or even 0 items.
func FetchAllPages[T any](ctx context.Context, fetcher ItemFetcher[T], perPage int, opts ...PaginationOptions) ([]T, error) {
	var maxItems int
	if len(opts) > 0 {
		maxItems = opts[0].MaxItems
	}

	var allItems []T
	page := 1

	for {
		items, resp, err := fetcher(ctx, page, perPage)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, items...)

		if maxItems > 0 && len(allItems) >= maxItems {
			allItems = allItems[:maxItems]
			break
		}

		if resp == nil {
			break
		}

		linkHeader := resp.Header.Get("Link")
		if !HasNextPageLink(linkHeader) {
			break
		}

		page++
	}

	return allItems, nil
}

// HasNextPageLink checks if the provided HTTP Link header string contains rel="next".
func HasNextPageLink(linkHeader string) bool {
	if linkHeader == "" {
		return false
	}
	parts := strings.Split(linkHeader, ",")
	for _, part := range parts {
		if linkNextRegex.MatchString(strings.TrimSpace(part)) {
			return true
		}
	}
	return false
}

// GraphQLPageInfo represents GraphQL pagination metadata.
type GraphQLPageInfo struct {
	HasNextPage bool
	EndCursor   string
}

// GraphQLFetcher defines a function that fetches a page using GraphQL.
type GraphQLFetcher[T any] func(ctx context.Context, cursor string, perPage int) ([]T, GraphQLPageInfo, error)

// FetchAllPagesGraphQL iterates through GraphQL paginated queries according to pageInfo.HasNextPage.
func FetchAllPagesGraphQL[T any](ctx context.Context, fetcher GraphQLFetcher[T], perPage int, opts ...PaginationOptions) ([]T, error) {
	var maxItems int
	if len(opts) > 0 {
		maxItems = opts[0].MaxItems
	}

	var allItems []T
	cursor := ""

	for {
		items, pageInfo, err := fetcher(ctx, cursor, perPage)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, items...)

		if maxItems > 0 && len(allItems) >= maxItems {
			allItems = allItems[:maxItems]
			break
		}

		if !pageInfo.HasNextPage {
			break
		}

		cursor = pageInfo.EndCursor
	}

	return allItems, nil
}
