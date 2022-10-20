package main

import (
	"os"
	"regexp"

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
	Entropy     float32       `json:"Entropy"`
	Author      string        `json:"Author"`
	Email       string        `json:"Email"`
	Date        string        `json:"Date"`
	Message     string        `json:"Message"`
	Tags        []interface{} `json:"Tags"`
	RuleID      string        `json:"RuleID"`
}

type GitleaksRepoResult struct {
	Repository string
	Org        string
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
	IgnoreCommits        []string            `yaml:"ignore_commits"`
	ReposToIgnore        map[string][]string `yaml:"repo_ignore"`
	Output               ConfOutput          `yaml:"output"`
	r_ignore_map         map[string][]*regexp.Regexp
}
type ConfGithubConfig struct {
	OrgsToScan []string `yaml:"orgs_to_scan"`
	DaysToScan int      `yaml:"days_to_scan"`
}
type ConfOutput struct {
	Format string `yaml:"format"`
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
	// build the regex map
	c.buildIgnoreMap()
	return c, nil
}
func (c *Conf) buildIgnoreMap() {
	r_ignore_map := make(map[string][]*regexp.Regexp)
	for ri, expressions := range c.ReposToIgnore {
		// init the slice
		r_ignore_map[ri] = make([]*regexp.Regexp, 0)
		for _, expr := range expressions {
			re, err := regexp.Compile(expr)
			if err != nil {
				log.Warn().Err(err).Str("repo", ri).Str("expression", expr).Msg("failed to compile regex, skipping")
				continue
			}
			r_ignore_map[ri] = append(r_ignore_map[ri], re)
		}
	}
	c.r_ignore_map = r_ignore_map
}
