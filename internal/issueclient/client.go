package issueclient

import "context"

// ErrAlreadyExists is returned when an open issue for the movie already exists.
type ErrAlreadyExists struct {
	Movie string
}

func (e ErrAlreadyExists) Error() string {
	return "issue for movie " + e.Movie + " already exists"
}

// ErrIgnored is returned when a matching issue has the "ignored" label.
type ErrIgnored struct {
	Movie string
}

func (e ErrIgnored) Error() string {
	return "issue for movie " + e.Movie + " is ignored"
}

// IssueClient abstracts the operations needed to manage missing-movie issues.
type IssueClient interface {
	CheckForExistingMovieIssue(ctx context.Context, movie string) error
	AddMissingMovie(ctx context.Context, movie string) error
}
