package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config represents the YAML file described about labels and repos
type Config struct {
	Labels []Label `yaml:"labels"`
	Repos  Repos   `yaml:"repos"`
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

func (cfg Config) getDefinedLabel(name string) (Label, error) {
	for _, label := range cfg.Labels {
		if label.Name == name {
			return label, nil
		}
	}
	return Label{}, fmt.Errorf("%s: no such defined label in config YAML", name)
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
