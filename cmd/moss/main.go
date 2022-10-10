package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

func check_gitleaks_conf(gitleaks_path string) error {
	_, err := os.ReadFile(gitleaks_path)
	if err != nil {
		log.Fatal().Err(err).Str("path", gitleaks_path).Msg("failed to read gitleaks toml")
		return err
	}
	return nil
}

func scan_repo(repo, pat string) GitleaksRepoResult {
	result := GitleaksRepoResult{
		Repository: repo,
		URL:        "",    // TODO, fix this using github repo sdk html_url
		IsPrivate:  false, // TODO, fix this using github repo sdk
	}
	// run gitleaks
	return result
}

func get_org_repos(orgname, pat string, daysago int) ([]*github.Repository, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	//
	time_ago := time.Now().AddDate(0, 0, (-1 * daysago))
	org_repos := make([]*github.Repository, 0)
	page := 0
	for {
		opt := &github.RepositoryListByOrgOptions{Type: "all", Sort: "pushed", Direction: "desc", ListOptions: github.ListOptions{Page: page}}
		repos, _, err := client.Repositories.ListByOrg(context.Background(), "github", opt)
		if err != nil {
			log.Error().Err(err).Str("org", orgname).Msg("Error getting repositories from Github")
			return nil, err
		}
		saw_older := false
		for _, repo := range repos {
			if *repo.Archived {
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

func main() {
	// setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("logging setup")
	// load the config file
	var conf Conf
	conf.getConfig("./configs/conf.yml")
	// check the gitleaks.toml file exists and isn't empty
	check_gitleaks_conf("./configs/gitleaks.toml")
	// check the PAT exists for each org
	pats := make(map[string]string, 0)
	for _, org := range conf.GithubConfig.OrgsToScan {
		pat := os.Getenv(fmt.Sprintf("PAT_%s", org))
		if pat == "" {
			log.Error().Str("org", org).Msg("PAT for org doesn't exist. Skipping it")
			continue
		}
		pats[org] = pat
	}
	// foreach org, get the repos according to days_to_scan
	all_repos := make([]*github.Repository, 0)
	for org, pat := range pats {
		repos, err := get_org_repos(org, pat, conf.GithubConfig.DaysToScan)
		if err != nil {
			log.Error().Err(err).Str("org", org).Msg("Failed to get repos from org. Continuing")
			continue
		}
		all_repos = append(all_repos, repos...)
	}
	// and queue up the repos
}
