package main

import (
	"errors"
	"fmt"
	"io"
	"log"
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

	Defined Config
	Actual  Config
	GitHub  *github.Client
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
	c.Defined = cfg

	client, err := github.NewClient(c.Option.DryRun)
	if err != nil {
		return err
	}
	c.GitHub = client

	if err := c.Validate(); err != nil {
		// fmt.Fprintf(c.Stderr, "Note: %s\n", err)
		// return nil
		return err
	}

	if c.Option.Import {
		return c.Import()
	}

	return c.Sync()
}

// applyLabels creates/edits labels described in YAML
func (c *CLI) applyLabels(owner, repo string, label github.Label) error {
	ghLabel, err := c.GitHub.GetLabel(owner, repo, label)
	if err != nil {
		return c.GitHub.CreateLabel(owner, repo, label)
	}

	if ghLabel.Description != label.Description || ghLabel.Color != label.Color {
		return c.GitHub.EditLabel(owner, repo, label)
	}

	return nil
}

// deleteLabels deletes the label not described in YAML but exists on GitHub
func (c *CLI) deleteLabels(owner, repo string) error {
	labels, err := c.GitHub.ListLabels(owner, repo)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if c.Defined.checkIfRepoHasLabel(owner+"/"+repo, label.Name) {
			// no need to delete
			continue
		}
		err := c.GitHub.DeleteLabel(owner, repo, label)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CLI) syncLabels(repo github.Repo) error {
	slugs := strings.Split(repo.Name, "/")
	if len(slugs) != 2 {
		return fmt.Errorf("repository name %q is invalid", repo.Name)
	}
	for _, labelName := range repo.Labels {
		label, err := c.Defined.getDefinedLabel(labelName)
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

func checkLabelDuplication(labels []github.Label, target github.Label) bool {
	for _, label := range labels {
		if cmp.Equal(label, target) {
			return true
		}
	}
	return false
}

func (c *CLI) Validate() error {
	if len(c.Defined.Repos) == 0 {
		return fmt.Errorf("no repos found in %s", c.Option.Config)
	}

	var cfg Config
	for _, repo := range c.Defined.Repos {
		slugs := strings.Split(repo.Name, "/")
		if len(slugs) != 2 {
			// TODO: log
			continue
		}
		labels, err := c.GitHub.ListLabels(slugs[0], slugs[1])
		if err != nil {
			log.Printf("[ERROR] failed to fetch labels %s/%s: %v", slugs[0], slugs[1], err)
			continue
		}
		var ls []string
		for _, label := range labels {
			ls = append(ls, label.Name)
		}
		repo.Labels = ls
		cfg.Repos = append(cfg.Repos, repo)
		for _, label := range labels {
			if c.checkLabelDuplication(cfg.Labels, label) {
				log.Printf("[WARN] %s is duplicate, so skip to add", label.Name)
				continue
			}
			cfg.Labels = append(cfg.Labels, label)
		}
	}

	// used for Import func
	c.Actual = cfg

	if cmp.Equal(cfg, c.Defined) {
		return errors.New("existing labels and defined labels are the same")
	}

	return nil
}

func (c *CLI) Import() error {
	f, err := os.Create(c.Option.Config)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(&c.Actual)
}

// Sync syncs labels based on YAML
func (c *CLI) Sync() error {
	var eg errgroup.Group
	for _, repo := range c.Defined.Repos {
		repo := repo
		eg.Go(func() error {
			return c.syncLabels(repo)
		})
	}
	return eg.Wait()
}
