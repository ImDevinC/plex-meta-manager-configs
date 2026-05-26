package forgejo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ImDevinC/plex-meta-manager-configs/internal/issueclient"

	"code.gitea.io/sdk/gitea"
)

func newTestForgejoClient(t *testing.T, server *httptest.Server) *Client {
	client, err := gitea.NewClient(server.URL, gitea.SetToken("test-token"))
	if err != nil {
		t.Fatalf("failed to create gitea client: %v", err)
	}
	return &Client{
		client: client,
		owner:  "o",
		repo:   "r",
	}
}

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
					"id":     1,
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
					"id":    2,
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
					"id":    3,
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
					"id":    4,
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
				switch r.URL.Path {
				case "/api/v1/version":
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(map[string]any{"version": "1.20.0"}); err != nil {
						t.Fatalf("failed to encode version response: %v", err)
					}
					return
				case "/api/v1/repos/o/r/issues":
					if got := r.URL.Query().Get("state"); got != "all" {
						t.Fatalf("expected state=all, got %q", got)
					}
					if got := r.URL.Query().Get("limit"); got != "100" {
						t.Fatalf("expected limit=100, got %q", got)
					}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(tt.issues); err != nil {
						t.Fatalf("failed to encode response: %v", err)
					}
					return
				default:
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
			}))
			defer server.Close()

			client := newTestForgejoClient(t, server)
			err := client.CheckForExistingMovieIssue(context.Background(), "Inception")
			if ok := tt.errTypeCheck(err); !ok {
				t.Fatalf("unexpected error result: %v", err)
			}
		})
	}
}

func TestAddMissingMovie(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/version":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{"version": "1.20.0"}); err != nil {
				t.Fatalf("failed to encode version response: %v", err)
			}
			return
		case "/api/v1/repos/o/r/issues":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if got, ok := body["title"].(string); !ok || got != "Missing poster for movie Inception" {
				t.Fatalf("unexpected title: %v", body["title"])
			}
			if assignees, ok := body["assignees"].([]any); !ok || len(assignees) != 1 || assignees[0] != "o" {
				t.Fatalf("unexpected assignees: %v", body["assignees"])
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{"id": 42, "title": body["title"]}); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
			return
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestForgejoClient(t, server)
	if err := client.AddMissingMovie(context.Background(), "Inception"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
