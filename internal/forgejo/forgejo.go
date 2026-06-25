package forgejo

import (
	"context"
	"fmt"
	"strings"

	"ImDevinC/plex-meta-manager-configs/internal/issueclient"

	"code.gitea.io/sdk/gitea"
)

type Client struct {
	client   *gitea.Client
	owner    string
	repo     string
	assignee string
}

const issueTitleBase string = `Missing poster for movie %s`

func NewForgejoClient(ctx context.Context, url string, token string, owner string, repo string, assignee string) (issueclient.IssueClient, error) {
	client, err := gitea.NewClient(url, gitea.SetToken(token), gitea.SetContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create forgejo client: %w", err)
	}
	return &Client{
		client:   client,
		owner:    owner,
		repo:     repo,
		assignee: assignee,
	}, nil
}

func (c *Client) CheckForExistingMovieIssue(ctx context.Context, movie string) error {
	title := fmt.Sprintf(issueTitleBase, movie)
	opts := gitea.ListIssueOption{
		ListOptions: gitea.ListOptions{PageSize: 100, Page: 1},
		State:       gitea.StateAll,
	}
	for {
		issues, resp, err := c.client.ListRepoIssues(c.owner, c.repo, opts)
		if err != nil {
			return fmt.Errorf("failed to list existing issues. %w", err)
		}
		for _, i := range issues {
			if i.Title != title {
				continue
			}
			if hasLabel(i.Labels, "ignored") {
				return issueclient.ErrIgnored{Movie: movie}
			}
			if strings.EqualFold(string(i.State), "open") {
				return issueclient.ErrAlreadyExists{Movie: movie}
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func hasLabel(labels []*gitea.Label, want string) bool {
	for _, label := range labels {
		if strings.EqualFold(label.Name, want) {
			return true
		}
	}
	return false
}

func (c *Client) AddMissingMovie(ctx context.Context, movie string) error {
	_, _, err := c.client.CreateIssue(c.owner, c.repo, gitea.CreateIssueOption{
		Title:     fmt.Sprintf(issueTitleBase, movie),
		Assignees: []string{c.assignee},
	})
	if err != nil {
		return fmt.Errorf("failed to create issue. %w", err)
	}
	return nil
}
