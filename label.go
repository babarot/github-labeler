package main

import (
	"context"

	"github.com/google/go-github/github"
)

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

// Get gets GitHub labels
func (c *githubClient) GetLabel(owner, repo string, label Label) (Label, error) {
	ctx := context.Background()
	ghLabel, _, err := c.Labeler.GetLabel(ctx, owner, repo, label.Name)
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
func (c *githubClient) CreateLabel(owner, repo string, label Label) error {
	ctx := context.Background()
	ghLabel := &github.Label{
		Name:        github.String(label.Name),
		Description: github.String(label.Description),
		Color:       github.String(label.Color),
	}
	if len(label.PreviousName) > 0 {
		c.logger.Printf("rename %q in %s/%s to %q", label.PreviousName, owner, repo, label.Name)
		_, _, err := c.Labeler.EditLabel(ctx, owner, repo, label.PreviousName, ghLabel)
		return err
	}
	c.logger.Printf("create %q in %s/%s", label.Name, owner, repo)
	_, _, err := c.Labeler.CreateLabel(ctx, owner, repo, ghLabel)
	return err
}

// Edit edits GitHub labels
func (c *githubClient) EditLabel(owner, repo string, label Label) error {
	ctx := context.Background()
	ghLabel := &github.Label{
		Name:        github.String(label.Name),
		Description: github.String(label.Description),
		Color:       github.String(label.Color),
	}
	c.logger.Printf("edit %q in %s/%s", label.Name, owner, repo)
	_, _, err := c.Labeler.EditLabel(ctx, owner, repo, label.Name, ghLabel)
	return err
}

// List lists GitHub labels
func (c *githubClient) ListLabels(owner, repo string) ([]Label, error) {
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}
	var labels []Label
	for {
		ghLabels, resp, err := c.Labeler.ListLabels(ctx, owner, repo, opt)
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
func (c *githubClient) DeleteLabel(owner, repo string, label Label) error {
	ctx := context.Background()
	c.logger.Printf("delete %q in %s/%s", label.Name, owner, repo)
	_, err := c.Labeler.DeleteLabel(ctx, owner, repo, label.Name)
	return err
}
