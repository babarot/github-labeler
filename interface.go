package main

import (
	"io"
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

	GitHub *githubClient
	Config Manifest
}

type Option struct {
	DryRun  bool   `long:"dry-run" description:"Just dry run"`
	Config  string `short:"c" long:"config" description:"Path to YAML file that labels are defined" default:"labels.yaml"`
	Import  bool   `long:"import" description:"Import existing labels if enabled"`
	Version bool   `long:"version" description:"Show version"`
}
