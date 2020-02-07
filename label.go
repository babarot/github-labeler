package main

import (
	"context"

	"github.com/google/go-github/github"
)

// Get gets GitHub labels
func (g *LabelService) Get(owner, repo string, label Label) (Label, error) {
	ctx := context.Background()
	ghLabel, _, err := g.client.github.Issues.GetLabel(ctx, owner, repo, label.Name)
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
		_, _, err := g.client.github.Issues.EditLabel(ctx, owner, repo, label.PreviousName, ghLabel)
		return err
	}
	g.client.logger.Printf("create %q in %s/%s", label.Name, owner, repo)
	if g.client.dryRun {
		return nil
	}
	_, _, err := g.client.github.Issues.CreateLabel(ctx, owner, repo, ghLabel)
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
	_, _, err := g.client.github.Issues.EditLabel(ctx, owner, repo, label.Name, ghLabel)
	return err
}

// List lists GitHub labels
func (g *LabelService) List(owner, repo string) ([]Label, error) {
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}
	var labels []Label
	for {
		ghLabels, resp, err := g.client.github.Issues.ListLabels(ctx, owner, repo, opt)
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
	_, err := g.client.github.Issues.DeleteLabel(ctx, owner, repo, label.Name)
	return err
}
