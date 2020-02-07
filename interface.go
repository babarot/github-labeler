package main

import (
	"context"

	"github.com/google/go-github/github"
)

type Labeler interface {
	GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error)
	EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error)
	CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error)
	ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error)
	DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error)
}

type githubClientImpl struct {
	ghClient *github.Client
}

func (l githubClientImpl) GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error) {
	return l.ghClient.Issues.GetLabel(ctx, owner, repo, name)
}

func (l githubClientImpl) EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error) {
	return l.ghClient.Issues.EditLabel(ctx, owner, repo, name, label)
}

func (l githubClientImpl) CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	return l.ghClient.Issues.CreateLabel(ctx, owner, repo, label)
}

func (l githubClientImpl) ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	return l.ghClient.Issues.ListLabels(ctx, owner, repo, opt)
}

func (l githubClientImpl) DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error) {
	return l.ghClient.Issues.DeleteLabel(ctx, owner, repo, name)
}

type githubClientDryRun struct {
	ghClient *github.Client
}

func (l githubClientDryRun) GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error) {
	return l.ghClient.Issues.GetLabel(ctx, owner, repo, name)
}

func (l githubClientDryRun) EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error) {
	return nil, nil, nil
}

func (l githubClientDryRun) CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	return nil, nil, nil
}

func (l githubClientDryRun) ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	return l.ghClient.Issues.ListLabels(ctx, owner, repo, opt)
}

func (l githubClientDryRun) DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error) {
	return nil, nil
}
