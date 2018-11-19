package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

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
	PreviousName string `yaml:"previous_name"`
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

func loadManifest(path string) (Manifest, error) {
	var m Manifest
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return m, err
	}
	err = yaml.Unmarshal(buf, &m)
	return m, err
}

func (m Manifest) getDefinedLabel(name string) (Label, error) {
	for _, label := range m.Labels {
		if label.Name == name {
			return label, nil
		}
	}
	return Label{}, fmt.Errorf("%s: no such defined label in manifest YAML", name)
}

func (m Manifest) checkIfRepoHasLabel(repoName, labelName string) bool {
	var labels []string
	for _, repo := range m.Repos {
		if repo.Name == repoName {
			labels = repo.Labels
			break
		}
	}
	for _, label := range labels {
		if label == labelName {
			return true
		}
	}
	return false
}

type githubClient struct {
	*github.Client

	dryRun bool

	common service

	Label *LabelService
}

// LabelService handles communication with the label related
// methods of GitHub API
type LabelService service

type service struct {
	client *githubClient
}

// Labeler is label-maker instance
type Labeler struct {
	github   *githubClient
	manifest Manifest
}

// Get gets GitHub labels
func (g *LabelService) Get(owner, repo string, label Label) (Label, error) {
	ctx := context.Background()
	fetchedLabel, _, err := g.client.Issues.GetLabel(ctx, owner, repo, label.Name)
	if err != nil {
		return Label{}, err
	}
	return Label{
		Name:        *fetchedLabel.Name,
		Description: *fetchedLabel.Description,
		Color:       *fetchedLabel.Color,
	}, nil
}

// Create creates GitHub labels
func (g *LabelService) Create(owner, repo string, label Label) error {
	ctx := context.Background()
	ghLabel := &github.Label{
		Name:        github.String(label.Name),
		Description: github.String(label.Description),
		Color:       github.String(label.Color),
	}
	if len(label.PreviousName) > 0 {
		log.Printf("rename %q in %s/%s to %q", label.PreviousName, owner, repo, label.Name)
		if g.client.dryRun {
			return nil
		}
		_, _, err := g.client.Issues.EditLabel(ctx, owner, repo, label.PreviousName, ghLabel)
		return err
	}
	log.Printf("create %q in %s/%s", label.Name, owner, repo)
	if g.client.dryRun {
		return nil
	}
	_, _, err := g.client.Issues.CreateLabel(ctx, owner, repo, ghLabel)
	return err
}

// Edit edits GitHub labels
func (g *LabelService) Edit(owner, repo string, label Label) error {
	ctx := context.Background()
	ghLabel := &github.Label{
		Name:        github.String(label.Name),
		Description: github.String(label.Description),
		Color:       github.String(label.Color),
	}
	log.Printf("edit %q in %s/%s", label.Name, owner, repo)
	if g.client.dryRun {
		return nil
	}
	_, _, err := g.client.Issues.EditLabel(ctx, owner, repo, label.Name, ghLabel)
	return err
}

// List lists GitHub labels
func (g *LabelService) List(owner, repo string) ([]Label, error) {
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}
	var labels []Label
	for {
		ghLabels, resp, err := g.client.Issues.ListLabels(ctx, owner, repo, opt)
		if err != nil {
			return labels, err
		}
		for _, ghLabel := range ghLabels {
			description := ""
			if ghLabel.Description != nil {
				description = *ghLabel.Description
			}
			labels = append(labels, Label{
				Name:        *ghLabel.Name,
				Description: description,
				Color:       *ghLabel.Color,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return labels, nil
}

// Delete deletes GitHub labels
func (g *LabelService) Delete(owner, repo string, label Label) error {
	ctx := context.Background()
	log.Printf("delete %q in %s/%s", label.Name, owner, repo)
	if g.client.dryRun {
		return nil
	}
	_, err := g.client.Issues.DeleteLabel(ctx, owner, repo, label.Name)
	return err
}

// applyLabels creates/edits labels described in YAML
func (l Labeler) applyLabels(owner, repo string, label Label) error {
	fetchedLabel, err := l.github.Label.Get(owner, repo, label)
	if err != nil {
		return l.github.Label.Create(owner, repo, label)
	}

	if fetchedLabel.Description != label.Description || fetchedLabel.Color != label.Color {
		return l.github.Label.Edit(owner, repo, label)
	}

	return nil
}

// deleteLabels deletes the label not described in YAML but exists on GitHub
func (l Labeler) deleteLabels(owner, repo string) error {
	labels, err := l.github.Label.List(owner, repo)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if l.manifest.checkIfRepoHasLabel(owner+"/"+repo, label.Name) {
			// no need to delete
			continue
		}
		err := l.github.Label.Delete(owner, repo, label)
		if err != nil {
			return err
		}
	}

	return nil
}

// Sync syncs labels based on YAML
func (l Labeler) Sync(repo Repo) error {
	slugs := strings.Split(repo.Name, "/")
	if len(slugs) != 2 {
		return fmt.Errorf("repository name %q is invalid", repo.Name)
	}
	for _, labelName := range repo.Labels {
		label, err := l.manifest.getDefinedLabel(labelName)
		if err != nil {
			return err
		}
		err = l.applyLabels(slugs[0], slugs[1], label)
		if err != nil {
			return err
		}
	}
	return l.deleteLabels(slugs[0], slugs[1])
}

func newLabeler(configPath string, dryRun bool) (Labeler, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return Labeler{}, errors.New("GITHUB_TOKEN is missing")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	m, err := loadManifest(configPath)
	if err != nil {
		return Labeler{}, err
	}

	gc := &githubClient{
		Client: client,
		dryRun: dryRun,
	}
	gc.common.client = gc
	gc.Label = (*LabelService)(&gc.common)
	return Labeler{
		github:   gc,
		manifest: m,
	}, nil
}

func main() {
	var (
		manifest = flag.String("manifest", "labels.yaml", "YAML file to be described about labels and repos")
		dryRun   = flag.Bool("dry-run", false, "dry run flag")
	)
	flag.Parse()

	labeler, err := newLabeler(*manifest, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err.Error())
		os.Exit(1)
	}

	eg := errgroup.Group{}
	for _, repo := range labeler.manifest.Repos {
		repo := repo
		eg.Go(func() error {
			return labeler.Sync(repo)
		})
	}

	if err := eg.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err.Error())
		os.Exit(1)
	}
}
