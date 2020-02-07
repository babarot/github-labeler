package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	yaml "gopkg.in/yaml.v2"
)

type CLI struct {
	Stdout io.Writer
	Stderr io.Writer
	Option Option

	Labeler Labeler
	Logger  *log.Logger
	Config  Config
}

type Option struct {
	DryRun  bool   `long:"dry-run" description:"Just dry run"`
	Config  string `long:"config" description:"Path to YAML file that labels are defined" default:"labels.yaml"`
	Import  bool   `long:"import" description:"Import existing labels if enabled"`
	Version bool   `long:"version" description:"Show version"`
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

	cfg, err := loadConfig(c.Option.Config)
	if err != nil {
		return err
	}
	c.Config = cfg

	c.Labeler = githubClientImpl{client}
	c.Logger = log.New(os.Stdout, "labeler: ", log.Ldate|log.Ltime)

	if c.Option.DryRun {
		c.Labeler = githubClientDryRun{client}
		c.Logger.SetPrefix("labeler (dry-run): ")
	}

	if len(c.Config.Repos) == 0 {
		return fmt.Errorf("no repos found in %s", c.Option.Config)
	}

	actual := c.ActualConfig()
	if cmp.Equal(actual, c.Config) {
		// no need to sync
		c.Logger.Printf("Claimed config and actual config is the same")
		return nil
	}

	if c.Option.Import {
		f, err := os.Create(c.Option.Config)
		if err != nil {
			return err
		}
		defer f.Close()
		return yaml.NewEncoder(f).Encode(&actual)
	}

	eg := errgroup.Group{}
	for _, repo := range c.Config.Repos {
		repo := repo
		eg.Go(func() error {
			return c.Sync(repo)
		})
	}

	return eg.Wait()
}

// applyLabels creates/edits labels described in YAML
func (c *CLI) applyLabels(owner, repo string, label Label) error {
	ghLabel, err := c.GetLabel(owner, repo, label)
	if err != nil {
		return c.CreateLabel(owner, repo, label)
	}

	if ghLabel.Description != label.Description || ghLabel.Color != label.Color {
		return c.EditLabel(owner, repo, label)
	}

	return nil
}

// deleteLabels deletes the label not described in YAML but exists on GitHub
func (c *CLI) deleteLabels(owner, repo string) error {
	labels, err := c.ListLabels(owner, repo)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if c.Config.checkIfRepoHasLabel(owner+"/"+repo, label.Name) {
			// no need to delete
			continue
		}
		err := c.DeleteLabel(owner, repo, label)
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

func (c *CLI) ActualConfig() Config {
	var cfg Config
	for _, repo := range c.Config.Repos {
		slugs := strings.Split(repo.Name, "/")
		if len(slugs) != 2 {
			// TODO: handle error
			continue
		}
		labels, err := c.ListLabels(slugs[0], slugs[1])
		if err != nil {
			// TODO: handle error
			continue
		}
		var ls []string
		for _, label := range labels {
			ls = append(ls, label.Name)
		}
		repo.Labels = ls
		cfg.Repos = append(cfg.Repos, repo)
		cfg.Labels = append(cfg.Labels, labels...)
	}
	return cfg
}
