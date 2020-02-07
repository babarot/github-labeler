package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
)

// These variables are set in Goreleaser
var (
	Version  = "unset"
	Revision = "unset"
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

type CLI struct {
	Stdout io.Writer
	Stderr io.Writer
	Option Option

	Client *githubClient
	Config Manifest
}

type Option struct {
	DryRun  bool   `long:"dry-run" description:"Just dry run"`
	Config  string `short:"c" long:"config" description:"Just dry run" default:"labels.yaml"`
	Version bool   `long:"version" description:"Show version"`
}

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
	logger *log.Logger

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
	ghLabel, _, err := g.client.Issues.GetLabel(ctx, owner, repo, label.Name)
	if err != nil {
		return Label{}, err
	}
	return Label{
		Name:        ghLabel.GetName(),
		Description: ghLabel.GetDescription(),
		Color:       ghLabel.GetName(),
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
		g.client.logger.Printf("rename %q in %s/%s to %q", label.PreviousName, owner, repo, label.Name)
		if g.client.dryRun {
			return nil
		}
		_, _, err := g.client.Issues.EditLabel(ctx, owner, repo, label.PreviousName, ghLabel)
		return err
	}
	g.client.logger.Printf("create %q in %s/%s", label.Name, owner, repo)
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
	g.client.logger.Printf("edit %q in %s/%s", label.Name, owner, repo)
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
			labels = append(labels, Label{
				Name:        ghLabel.GetName(),
				Description: ghLabel.GetDescription(),
				Color:       ghLabel.GetColor(),
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
	g.client.logger.Printf("delete %q in %s/%s", label.Name, owner, repo)
	if g.client.dryRun {
		return nil
	}
	_, err := g.client.Issues.DeleteLabel(ctx, owner, repo, label.Name)
	return err
}

// applyLabels creates/edits labels described in YAML
func (c *CLI) applyLabels(owner, repo string, label Label) error {
	ghLabel, err := c.Client.Label.Get(owner, repo, label)
	if err != nil {
		return c.Client.Label.Create(owner, repo, label)
	}

	if ghLabel.Description != label.Description || ghLabel.Color != label.Color {
		return c.Client.Label.Edit(owner, repo, label)
	}

	return nil
}

// deleteLabels deletes the label not described in YAML but exists on GitHub
func (c *CLI) deleteLabels(owner, repo string) error {
	labels, err := c.Client.Label.List(owner, repo)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if c.Config.checkIfRepoHasLabel(owner+"/"+repo, label.Name) {
			// no need to delete
			continue
		}
		err := c.Client.Label.Delete(owner, repo, label)
		if err != nil {
			return err
		}
	}

	return nil
}

// Sync syncs labels based on YAML
func (c *CLI) Sync(repo Repo) error {
	slugs := strings.Split(repo.Name, "/")
	if len(slugs) != 2 {
		return fmt.Errorf("repository name %q is invalid", repo.Name)
	}
	for _, labelName := range repo.Labels {
		label, err := c.Config.getDefinedLabel(labelName)
		if err != nil {
			return err
		}
		err = c.applyLabels(slugs[0], slugs[1], label)
		if err != nil {
			return err
		}
	}
	return c.deleteLabels(slugs[0], slugs[1])
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// clilog.Env = "GOMI_LOG"
	// clilog.SetOutput()
	// defer log.Printf("[INFO] finish main function")
	//
	// log.Printf("[INFO] Version: %s (%s)", Version, Revision)
	// log.Printf("[INFO] gomiPath: %s", gomiPath)
	// log.Printf("[INFO] inventoryPath: %s", inventoryPath)
	// log.Printf("[INFO] Args: %#v", args)

	var opt Option
	args, err := flags.ParseArgs(&opt, args)
	if err != nil {
		return 2
	}

	cli := CLI{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Option: opt,
	}

	if err := cli.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return 0
}

func (c *CLI) Run(args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN is missing")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	m, err := loadManifest(c.Option.Config)
	if err != nil {
		return err
	}

	gc := &githubClient{
		Client: client,
		dryRun: c.Option.DryRun,
		logger: log.New(os.Stdout, "labeler: ", log.Ldate|log.Ltime),
	}

	if c.Option.DryRun {
		gc.logger.SetPrefix("labeler (dry-run): ")
	}

	gc.common.client = gc
	gc.Label = (*LabelService)(&gc.common)

	c.Client = gc
	c.Config = m

	eg := errgroup.Group{}
	for _, repo := range c.Config.Repos {
		repo := repo
		eg.Go(func() error {
			return c.Sync(repo)
		})
	}

	return eg.Wait()
}
