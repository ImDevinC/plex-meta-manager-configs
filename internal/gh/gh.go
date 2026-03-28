package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

type ErrAlreadyExists struct {
	Movie string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("issue for movie %s already exists", e.Movie)
}

type ErrIgnored struct {
	Movie string
}

func (e ErrIgnored) Error() string {
	return fmt.Sprintf("issue for movie %s is ignored", e.Movie)
}

type Client struct {
	githubClient *github.Client
	owner        string
	repo         string
}

const issueTitleBase string = `Missing poster for movie %s`

func NewClient(ctx context.Context, accessToken string, owner string, repo string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		githubClient: github.NewClient(tc),
		owner:        owner,
		repo:         repo,
	}
}

func (c *Client) CheckForExistingMovieIssue(ctx context.Context, movie string) error {
	title := fmt.Sprintf(issueTitleBase, movie)
	opts := &github.IssueListByRepoOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		issues, resp, err := c.githubClient.Issues.ListByRepo(ctx, c.owner, c.repo, opts)
		if err != nil {
			return fmt.Errorf("failed to list existing issues. %w", err)
		}
		for _, i := range issues {
			if i.Title == nil || *i.Title != title {
				continue
			}
			if hasLabel(i.Labels, "ignored") {
				return ErrIgnored{Movie: movie}
			}
			if strings.EqualFold(i.GetState(), "open") {
				return ErrAlreadyExists{Movie: movie}
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func hasLabel(labels []*github.Label, want string) bool {
	for _, label := range labels {
		if strings.EqualFold(label.GetName(), want) {
			return true
		}
	}
	return false
}

func (c *Client) AddMissingMovie(ctx context.Context, movie string) error {
	request := &github.IssueRequest{
		Title:    ptr.String(fmt.Sprintf(issueTitleBase, movie)),
		Assignee: ptr.String(c.owner),
	}
	_, _, err := c.githubClient.Issues.Create(ctx, c.owner, c.repo, request)
	if err != nil {
		return fmt.Errorf("failed to create issue. %w", err)
	}

	return nil
}
