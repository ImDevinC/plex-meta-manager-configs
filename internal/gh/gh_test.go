package gh

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v43/github"
)

func TestCheckForExistingMovieIssue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		issues       []map[string]any
		expectErr    bool
		errTypeCheck func(error) bool
	}{
		{
			name: "returns ignored when matching issue has ignored label",
			issues: []map[string]any{
				{
					"title":  "Missing poster for movie Inception",
					"state":  "closed",
					"labels": []map[string]any{{"name": "ignored"}},
				},
			},
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var ignoredErr ErrIgnored
				return err != nil && errors.As(err, &ignoredErr)
			},
		},
		{
			name: "returns already exists when matching issue is open",
			issues: []map[string]any{
				{
					"title": "Missing poster for movie Inception",
					"state": "open",
				},
			},
			expectErr: true,
			errTypeCheck: func(err error) bool {
				var existsErr ErrAlreadyExists
				return err != nil && errors.As(err, &existsErr)
			},
		},
		{
			name: "returns nil when matching issue is closed and not ignored",
			issues: []map[string]any{
				{
					"title": "Missing poster for movie Inception",
					"state": "closed",
				},
			},
			expectErr: false,
			errTypeCheck: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "returns nil when no matching issue exists",
			issues: []map[string]any{
				{
					"title": "Missing poster for movie Interstellar",
					"state": "open",
				},
			},
			expectErr: false,
			errTypeCheck: func(err error) bool {
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/o/r/issues" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("state"); got != "all" {
					t.Fatalf("expected state=all, got %q", got)
				}
				if got := r.URL.Query().Get("per_page"); got != "100" {
					t.Fatalf("expected per_page=100, got %q", got)
				}

				if err := json.NewEncoder(w).Encode(tt.issues); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			ghClient := github.NewClient(server.Client())
			baseURL, err := url.Parse(server.URL + "/")
			if err != nil {
				t.Fatalf("failed to parse test server URL: %v", err)
			}
			ghClient.BaseURL = baseURL

			client := &Client{
				githubClient: ghClient,
				owner:        "o",
				repo:         "r",
			}

			err = client.CheckForExistingMovieIssue(context.Background(), "Inception")
			if ok := tt.errTypeCheck(err); !ok {
				t.Fatalf("unexpected error result: %v", err)
			}
		})
	}
}
