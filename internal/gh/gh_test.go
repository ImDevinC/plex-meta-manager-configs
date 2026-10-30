package gh

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ImDevinC/plex-meta-manager-configs/internal/issueclient"

	"github.com/google/go-github/v93/github"
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
				var ignoredErr issueclient.ErrIgnored
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
				var existsErr issueclient.ErrAlreadyExists
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

			baseURL := server.URL + "/"
			ghClient, err := github.NewClient(
				github.WithHTTPClient(server.Client()),
				github.WithURLs(&baseURL, nil),
			)
			if err != nil {
				t.Fatalf("failed to create GitHub client: %v", err)
			}

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

func TestAddMissingMovie(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/repos/o/r/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var request struct {
			Title    string `json:"title"`
			Assignee string `json:"assignee"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if request.Title != "Missing poster for movie Inception" {
			t.Errorf("unexpected issue title: %q", request.Title)
		}
		if request.Assignee != "o" {
			t.Errorf("unexpected assignee: %q", request.Assignee)
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	baseURL := server.URL + "/"
	ghClient, err := github.NewClient(
		github.WithHTTPClient(server.Client()),
		github.WithURLs(&baseURL, nil),
	)
	if err != nil {
		t.Fatalf("failed to create GitHub client: %v", err)
	}

	client := &Client{
		githubClient: ghClient,
		owner:        "o",
		repo:         "r",
	}
	if err := client.AddMissingMovie(context.Background(), "Inception"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
