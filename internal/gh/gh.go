package gh

import (
	"context"
	"fmt"

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
		State: "open",
	}
	issues, _, err := c.githubClient.Issues.ListByRepo(ctx, c.owner, c.repo, opts)
	if err != nil {
		return fmt.Errorf("failed to list existing issues. %w", err)
	}
	for _, i := range issues {
		if *i.Title == title {
			return ErrAlreadyExists{Movie: movie}
		}
	}
	return nil
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
