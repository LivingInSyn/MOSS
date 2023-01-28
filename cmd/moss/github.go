package main

import (
	"context"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// Fetch all the github repos
func get_all_github_repos(pats map[string]string, conf Conf) map[string]*GitRepo {
	all_repos := make(map[string]*GitRepo, 0)
	for org, pat := range pats {
		repos, err := get_org_repos(org, pat, conf.GithubConfig.DaysToScan, conf.SkipRepos)
		if err != nil {
			log.Error().Err(err).Str("org", org).Msg("Failed to get repos from org. Continuing")
			continue
		}
		for _, repo := range repos {
			all_repos[*repo.HTMLURL] = github_to_git(repo)
			all_repos[*repo.HTMLURL].pat = pat
			all_repos[*repo.HTMLURL].provider = "GITHUB"

		}
	}
	return all_repos
}

// Get github repos for the respective ORGs
func get_org_repos(orgname, pat string, daysago int, skipRepos []string) ([]*github.Repository, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	//TEMP
	// r, _, _ := client.Repositories.Get(context.Background(), "puppetlabs", "puppetlabs-docker")
	// return []*github.Repository{r}, nil
	//end temp

	time_ago := time.Now().AddDate(0, 0, (-1 * daysago))
	org_repos := make([]*github.Repository, 0)
	page := 0
	for {
		opt := &github.RepositoryListByOrgOptions{Type: "all", Sort: "pushed", Direction: "desc", ListOptions: github.ListOptions{Page: page}}
		repos, _, err := client.Repositories.ListByOrg(context.Background(), orgname, opt)
		if err != nil {
			log.Error().Err(err).Str("org", orgname).Msg("Error getting repositories from Github")
			return nil, err
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
			if repo.PushedAt.Time.Before(time_ago) {
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
