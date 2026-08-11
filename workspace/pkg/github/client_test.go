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
