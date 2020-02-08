package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/b4b4r07/github-labeler/pkg/github"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
	yaml "gopkg.in/yaml.v2"
)

type CLI struct {
	Stdout io.Writer
	Stderr io.Writer
	Option Option

	Config Config
	Client *github.Client
}

type Option struct {
	DryRun  bool   `long:"dry-run" description:"Just dry run"`
	Config  string `long:"config" description:"Path to YAML file that labels are defined" default:"labels.yaml"`
	Import  bool   `long:"import" description:"Import existing labels if enabled"`
	Version bool   `long:"version" description:"Show version"`
}

func (c *CLI) Run(args []string) error {
	if c.Option.Version {
		fmt.Fprintf(c.Stdout, "%s (%s)\n", Version, Revision)
		return nil
	}

	cfg, err := loadConfig(c.Option.Config)
	if err != nil {
		return err
	}
	c.Config = cfg

	client, err := github.NewClient(c.Option.DryRun)
	if err != nil {
		return err
	}
	c.Client = client

	if len(c.Config.Repos) == 0 {
		return fmt.Errorf("no repos found in %s", c.Option.Config)
	}

	actual := c.ActualConfig()
	if cmp.Equal(actual, c.Config) {
		fmt.Fprintf(c.Stdout, "no need to sync (actual and desired is the same)\n")
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
func (c *CLI) applyLabels(owner, repo string, label github.Label) error {
	ghLabel, err := c.Client.GetLabel(owner, repo, label)
	if err != nil {
		return c.Client.CreateLabel(owner, repo, label)
	}

	if ghLabel.Description != label.Description || ghLabel.Color != label.Color {
		return c.Client.EditLabel(owner, repo, label)
	}

	return nil
}

// deleteLabels deletes the label not described in YAML but exists on GitHub
func (c *CLI) deleteLabels(owner, repo string) error {
	labels, err := c.Client.ListLabels(owner, repo)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if c.Config.checkIfRepoHasLabel(owner+"/"+repo, label.Name) {
			// no need to delete
			continue
		}
		err := c.Client.DeleteLabel(owner, repo, label)
		if err != nil {
			return err
		}
	}

	return nil
}

// Sync syncs labels based on YAML
func (c *CLI) Sync(repo github.Repo) error {
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
		labels, err := c.Client.ListLabels(slugs[0], slugs[1])
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
