package main

import (
	"fmt"
	"io/ioutil"

	"github.com/b4b4r07/github-labeler/pkg/github"
	yaml "gopkg.in/yaml.v2"
)

// Config represents the YAML file described about labels and repos
type Config struct {
	Labels []github.Label `yaml:"labels"`
	Repos  github.Repos   `yaml:"repos"`
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(buf, &cfg)
	return cfg, err
}

func (cfg Config) getDefinedLabel(name string) (github.Label, error) {
	for _, label := range cfg.Labels {
		if label.Name == name {
			return label, nil
		}
	}
	return github.Label{}, fmt.Errorf("%s: no such defined label in config YAML", name)
}

func (cfg Config) checkIfRepoHasLabel(repoName, labelName string) bool {
	var labels []string
	for _, repo := range cfg.Repos {
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
