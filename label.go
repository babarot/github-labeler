package main

import (
	"context"

	"github.com/google/go-github/github"
)

// Get gets GitHub labels
func (c *githubClient) GetLabel(owner, repo string, label Label) (Label, error) {
	ctx := context.Background()
	ghLabel, _, err := c.Issues.GetLabel(ctx, owner, repo, label.Name)
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
		if c.dryRun {
			return nil
		}
		_, _, err := c.Issues.EditLabel(ctx, owner, repo, label.PreviousName, ghLabel)
		return err
	}
	c.logger.Printf("create %q in %s/%s", label.Name, owner, repo)
	if c.dryRun {
		return nil
	}
	_, _, err := c.Issues.CreateLabel(ctx, owner, repo, ghLabel)
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
	if c.dryRun {
		return nil
	}
	_, _, err := c.Issues.EditLabel(ctx, owner, repo, label.Name, ghLabel)
	return err
}

// List lists GitHub labels
func (c *githubClient) ListLabels(owner, repo string) ([]Label, error) {
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}
	var labels []Label
	for {
		ghLabels, resp, err := c.Issues.ListLabels(ctx, owner, repo, opt)
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
	if c.dryRun {
		return nil
	}
	_, err := c.Issues.DeleteLabel(ctx, owner, repo, label.Name)
	return err
}
