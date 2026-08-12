package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issues := []Issue{
			{ID: 1, Title: "Issue 1"},
			{ID: 2, Title: "Issue 2"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issues)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	issues, err := client.GetIssues("owner", "repo", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestGetIssues_SparsePagination(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount == 1 {
			issues := []Issue{
				{ID: 1, Title: "Issue 1"},
				{ID: 2, Title: "Issue 2"},
			}
			w.Header().Set("Link", `<`+server.URL+`/repos/owner/repo/issues?page=2&per_page=5>; rel="next"`)
			json.NewEncoder(w).Encode(issues)
		} else if requestCount == 2 {
			issues := []Issue{
				{ID: 3, Title: "Issue 3"},
				{ID: 4, Title: "Issue 4"},
				{ID: 5, Title: "Issue 5"},
			}
			json.NewEncoder(w).Encode(issues)
		} else {
			t.Errorf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	issues, err := client.GetIssues("owner", "repo", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected exactly 2 requests, got %d", requestCount)
	}

	if len(issues) != 5 {
		t.Errorf("expected 5 issues, got %d", len(issues))
	}
}

func TestGetIssues_Limit(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount == 1 {
			issues := []Issue{
				{ID: 1, Title: "Issue 1"},
				{ID: 2, Title: "Issue 2"},
				{ID: 3, Title: "Issue 3"},
				{ID: 4, Title: "Issue 4"},
			}
			w.Header().Set("Link", `<`+server.URL+`/repos/owner/repo/issues?page=2&per_page=3>; rel="next"`)
			json.NewEncoder(w).Encode(issues)
		} else {
			t.Errorf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	issues, err := client.GetIssues("owner", "repo", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("expected exactly 1 request (limit reached on first page), got %d", requestCount)
	}

	if len(issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(issues))
	}
}

func TestGetIssues_EmptyPageStops(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		// Empty page with a next link must NOT be followed — the empty-list
		// fallback stops pagination to prevent infinite loops.
		w.Header().Set("Link", `<`+server.URL+`/repos/owner/repo/issues?page=2&per_page=3>; rel="next"`)
		json.NewEncoder(w).Encode([]Issue{})
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	issues, err := client.GetIssues("owner", "repo", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("expected exactly 1 request (empty page stops pagination), got %d", requestCount)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}
