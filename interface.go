package main

import (
	"context"
	"io"
	"log"

	"github.com/google/go-github/github"
)

// Manifest represents the YAML file described about labels and repos
type Manifest struct {
	Labels Labels `yaml:"labels"`
	Repos  Repos  `yaml:"repos"`
}

// Label represents GitHub label
type Label struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Color        string `yaml:"color"`
	PreviousName string `yaml:"previous_name,omitempty"`
}

// Labels represents a collection of Label
type Labels []Label

// Repo represents GitHub repository
type Repo struct {
	Name   string   `yaml:"name"`
	Labels []string `yaml:"labels"`
}

// Repos represents a collection of Repo
type Repos []Repo

type CLI struct {
	Stdout io.Writer
	Stderr io.Writer
	Option Option

	Client *githubClient
	Config Manifest
}

type Option struct {
	DryRun  bool   `long:"dry-run" description:"Just dry run"`
	Config  string `short:"c" long:"config" description:"Path to YAML file that labels are defined" default:"labels.yaml"`
	Import  bool   `long:"import" description:"Import existing labels if enabled"`
	Version bool   `long:"version" description:"Show version"`
}

type githubClient struct {
	// github *github.Client
	github *github.Client

	dryRun bool
	logger *log.Logger

	common service

	Label *LabelService
}

type Labeler interface {
	// Get(owner, repo string, label Label) (Label, error)
	// Create(owner, repo string, label Label) error
	// Edit(owner, repo string, label Label) error
	// List(owner, repo string) ([]Label, error)
	// Delete(owner, repo string, label Label) error

	GetLabel(ctx context.Context, owner string, repo string, name string) (*github.Label, *github.Response, error)
	EditLabel(ctx context.Context, owner string, repo string, name string, label *github.Label) (*github.Label, *github.Response, error)
	CreateLabel(ctx context.Context, owner string, repo string, label *github.Label) (*github.Label, *github.Response, error)
	ListLabels(ctx context.Context, owner string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error)
	DeleteLabel(ctx context.Context, owner string, repo string, name string) (*github.Response, error)
}

// LabelService handles communication with the label related
// methods of GitHub API
type LabelService service

type service struct {
	client *githubClient
}
