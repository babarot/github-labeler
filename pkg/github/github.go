package github

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Client struct {
	Labeler Labeler
	Logger  *log.Logger
}

func NewClient(dryrun bool) (*Client, error) {
	labeler, err := newGitHubClient(dryrun)
	if err != nil {
		return nil, err
	}

	logger := log.New(os.Stdout, "labeler: ", log.Ldate|log.Ltime)
	if dryrun {
		logger.SetPrefix("labeler (dry-run): ")
	}

	return &Client{
		Labeler: labeler,
		Logger:  logger,
	}, nil
}

type Labeler interface {
	GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error)
	EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error)
	CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error)
	ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error)
	DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error)
}

func newGitHubClient(dryrun bool) (Labeler, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN is missing")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	if dryrun {
		return githubClientDryRun{client}, nil
	}

	return githubClientImpl{client}, nil
}

type githubClientImpl struct {
	*github.Client
}

func (l githubClientImpl) GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error) {
	return l.Issues.GetLabel(ctx, owner, repo, name)
}

func (l githubClientImpl) EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error) {
	return l.Issues.EditLabel(ctx, owner, repo, name, label)
}

func (l githubClientImpl) CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	return l.Issues.CreateLabel(ctx, owner, repo, label)
}

func (l githubClientImpl) ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	return l.Issues.ListLabels(ctx, owner, repo, opt)
}

func (l githubClientImpl) DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error) {
	return l.Issues.DeleteLabel(ctx, owner, repo, name)
}

type githubClientDryRun struct {
	*github.Client
}

func (l githubClientDryRun) GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error) {
	return l.Issues.GetLabel(ctx, owner, repo, name)
}

func (l githubClientDryRun) EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error) {
	return nil, nil, nil
}

func (l githubClientDryRun) CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	return nil, nil, nil
}

func (l githubClientDryRun) ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	return l.Issues.ListLabels(ctx, owner, repo, opt)
}

func (l githubClientDryRun) DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error) {
	return nil, nil
}
