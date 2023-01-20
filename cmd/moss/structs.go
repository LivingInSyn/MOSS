package main

import (
	"os"
	"regexp"
	"time"

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

type GitRepo struct {
	Name     string
	FullName string
	CloneURL string
	HTMLURL  string
	Private  bool
	orgname  string
	Archived bool
	PushedAt time.Time
	pat      string
	provider string
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
	GitlabConfig         ConfGitlabConfig    `yaml:"gitlab_config"`
	GithubConfig         ConfGithubConfig    `yaml:"github_config"`
	GitLeaksConfig       GitLeaksConfig      `yaml:"gitleaks_config"`
	SkipRepos            []string            `yaml:"skip_repos"`
	IgnoreSecretPatterns []string            `yaml:"ignore_secret_pattern"`
	IgnoreSecrets        []string            `yaml:"ignore_secrets"`
	IgnoreCommits        []string            `yaml:"ignore_commits"`
	ReposToIgnore        map[string][]string `yaml:"repo_ignore"`
	Output               ConfOutput          `yaml:"output"`
	// r_ignore_map is the ignoring of paths in repos
	r_ignore_map map[string][]*regexp.Regexp
	// s_ignores is the slice of regular expressions for secrets to ignore
	s_ignores []*regexp.Regexp
}
type ConfGithubConfig struct {
	OrgsToScan []string `yaml:"orgs_to_scan"`
	DaysToScan int      `yaml:"days_to_scan"`
}
type ConfGitlabConfig struct {
	OrgsToScan []string `yaml:"orgs_to_scan"`
	DaysToScan int      `yaml:"days_to_scan"`
}
type GitLeaksConfig struct {
	AdditionalArgs []string `yaml:"additional_args"`
}
type ConfOutput struct {
	Format string `yaml:"format"`
}
type RepoScanResult struct {
	Repository string
	URL        string
	IsPrivate  bool
	Org        string
	ERR        error
}

func (c *Conf) getConfig(confPath string) (*Conf, error) {
	yamlFile, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatal().Err(err).Str("confPath", confPath).Msg("Failed to read config file")
		return &Conf{}, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config file")
		return &Conf{}, err
	}
	// build the regex map
	c.buildIgnoreMap()
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
func (c *Conf) buildSecretIgnores() {
	ignore_patterns := make([]*regexp.Regexp, 0)
	for _, expr := range c.IgnoreSecretPatterns {
		re, err := regexp.Compile(expr)
		if err != nil {
			log.Warn().Err(err).Str("expr", expr).Msg("Ignore Secret Pattern is invalid. Continuing without it")
			continue
		}
		ignore_patterns = append(ignore_patterns, re)
	}
	c.s_ignores = ignore_patterns
}
