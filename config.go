package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	version         = "0.1.0"
	DefaultLastDays = 15
)

// Config is the structured configuration for cloud-console.yaml.
type Config struct {
	GithubToken string `json:"github-token" yaml:"github-token"`
	// StartDate is start time of statistics.
	StartDate string `json:"start-date" yaml:"start-date"`
	// EndDate is end time of statistics.
	EndDate string `json:"end-date" yaml:"end-date"`

	// LastDays is the past days to statistic.
	LastDays int `json:"last-days" yaml:"last-days"`
	// ToCurrent is whether to statistic to current time, otherwise to 23:59:59 of the previous day
	ToCurrent bool `json:"to-current" yaml:"to-current"`

	// Repositories is repositories to statistic
	Repositories []string `json:"repositories" yaml:"repositories"`
}

// NewFromFile parses the content in the file and creates a Config object.
func NewFromFile(filepath string) (*Config, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
