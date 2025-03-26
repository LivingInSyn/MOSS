package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// Convert Github to common Git Repo Struct
func github_to_git(project *github.Repository, pat string) *GitRepo {
	return &GitRepo{
		Name:     project.GetName(),
		FullName: project.GetFullName(),
		CloneURL: project.GetCloneURL(),
		HTMLURL:  project.GetHTMLURL(),
		Private:  project.GetPrivate(),
		orgname:  project.GetOwner().GetLogin(),
		Archived: project.GetArchived(),
		PushedAt: project.GetPushedAt().Time,
		pat:      pat,
		provider: "GITHUB",
	}
}

func getPat(provider string, org OrgConfig) string {
	token := ""
	provider = strings.ToUpper(provider)
	if org.Type == "cloud" {
		token = os.Getenv(provider + "_PAT_CLOUD_" + org.Name)
	}
	if org.Type == "onprem" {
		token = os.Getenv(provider + "_PAT_ONPREM_" + org.Name)
	}
	if token == "" {
		log.Error().Str("org", org.Name).Msg("token missing for org")
	}
	return token
}

func InitGitHubClient(org OrgConfig, token string) (*github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	if org.Type == "onprem" {
		if org.BaseURL == "" {
			return nil, fmt.Errorf("GitHub on-prem org '%s' requires base_url", org.Name)
		}
		client, err := github.NewEnterpriseClient(org.BaseURL, org.BaseURL, tc)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	return github.NewClient(tc), nil
}

func get_all_github_repos(orgs []OrgConfig, conf Conf) map[string]*GitRepo {
	all_repos := make(map[string]*GitRepo, 0)

	for _, org := range orgs {
		pat := getPat("github", org)
		client, err := InitGitHubClient(org, pat)
		if err != nil {
			log.Printf("[GitHub:%s] client init error: %v", org.Name, err)
			continue
		}
		log.Info().Str("org", org.Name).Str("type", org.Type).Msg("connected to GitHub")
		repos, err := get_org_repos(org, client, conf.GithubConfig.DaysToScan, conf.SkipRepos)

		if err != nil {
			log.Error().Err(err).Str("org", org.Name).Msg("Failed to get repos from org. Continuing")
			continue
		}

		for _, repo := range repos {
			all_repos[*repo.HTMLURL] = github_to_git(repo, pat)
		}
	}
	return all_repos
}

// Get github repos for the respective ORGs
func get_org_repos(org OrgConfig, client *github.Client, daysago int, skipRepos []string) ([]*github.Repository, error) {
	time_ago := time.Now().AddDate(0, 0, (-1 * daysago))
	org_repos := make([]*github.Repository, 0)
	page := 1
	for {
		opt := &github.RepositoryListByOrgOptions{Type: "all", Sort: "pushed", Direction: "desc", ListOptions: github.ListOptions{Page: page}}
		repos, _, err := client.Repositories.ListByOrg(context.Background(), org.Name, opt)
		if err != nil {
			log.Error().Err(err).Str("org", org.Name).Msg("Error getting repositories from Github")
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		saw_older := false
		for _, repo := range repos {

			if *repo.Archived {
				log.Debug().Str("repo", *repo.FullName).Msg("skipping repo because it's archived")
				continue
			}
			if skip_repo(repo, skipRepos) {
				log.Debug().Str("repo", *repo.FullName).Msg("skipping repo due to config")
				continue
			}
			if daysago > 0 && repo.PushedAt.Time.Before(time_ago) {
				saw_older = true
				break
			}
			org_repos = append(org_repos, repo)
		}
		if saw_older {
			break
		}
		page = page + 1
	}
	return org_repos, nil
}
