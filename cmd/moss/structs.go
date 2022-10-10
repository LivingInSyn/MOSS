package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type GitleaksResult struct {
	Description string        `json:"Description"`
	StartLine   int           `json:"StartLine"`
	EndLine     int           `json:"EndLine"`
	StartColumn int           `json:"StartColumn"`
	EndColumn   int           `json:"EndColumn"`
	Match       string        `json:"Match"`
	Secret      string        `json:"Secret"`
	File        string        `json:"File"`
	Commit      string        `json:"Commit"`
	Entropy     int           `json:"Entropy"`
	Author      string        `json:"Author"`
	Email       string        `json:"Email"`
	Date        string        `json:"Date"`
	Message     string        `json:"Message"`
	Tags        []interface{} `json:"Tags"`
	RuleID      string        `json:"RuleID"`
}

type GitleaksRepoResult struct {
	Repository string
	URL        string
	Err        error
	IsPrivate  bool
	Results    []GitleaksResult
}

type Conf struct {
	GithubConfig         ConfGithubConfig    `yaml:"github_config"`
	SkipRepos            []string            `yaml:"skip_repos"`
	IgnoreSecretPatterns []string            `yaml:"ignore_secret_pattern"`
	IgnoreSecrets        []string            `yaml:"ignore_secrets"`
	ReposToIgnore        map[string][]string `yaml:"repo_ignore"`
}
type ConfGithubConfig struct {
	OrgsToScan []string `yaml:"orgs_to_scan"`
	DaysToScan int      `yaml:"days_to_scan"`
}

func (c *Conf) getConfig(confPath string) (*Conf, error) {
	yamlFile, err := os.ReadFile(confPath)
	if err != nil {
		log.Error().Err(err).Str("confPath", confPath).Msg("Failed to read config file")
		return &Conf{}, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal config file")
		return &Conf{}, err
	}
	return c, nil
}
