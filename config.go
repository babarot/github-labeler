package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Manifest represents the YAML file described about labels and repos
type Manifest struct {
	Labels Labels `yaml:"labels"`
	Repos  Repos  `yaml:"repos"`
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
