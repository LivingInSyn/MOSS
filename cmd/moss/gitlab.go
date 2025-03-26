package main

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xanzy/go-gitlab"
)

// Helper function for skipping the repos
func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

// Convert Gitlab to common Git Repo Struct
func gitlab_to_git(p *gitlab.Project, pat, org string) *GitRepo {
	return &GitRepo{
		Name:     p.Name,
		FullName: p.PathWithNamespace,
		CloneURL: p.HTTPURLToRepo,
		HTMLURL:  p.WebURL,
		Private:  p.Visibility == "private",
		Archived: p.Archived,
		PushedAt: *p.LastActivityAt,
		pat:      pat,
		orgname:  org,
		provider: "GITLAB",
	}
}

func InitGitLabClient(org OrgConfig, token string) (*gitlab.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitLab token missing for org: %s", org.Name)
	}
	if org.Type == "onprem" {
		if org.BaseURL == "" {
			return nil, fmt.Errorf("GitLab on-prem org '%s' requires base_url", org.Name)
		}
		return gitlab.NewClient(token, gitlab.WithBaseURL(org.BaseURL))
	}
	return gitlab.NewClient(token)
}

func get_all_gitlab_repos(orgs []OrgConfig, conf Conf) map[string]*GitRepo {
	gitlab_repos := make(map[string]*GitRepo)
	time_ago := time.Now().AddDate(0, 0, (-1 * conf.GitlabConfig.DaysToScan))
	for _, org := range orgs {
		git, err := InitGitLabClient(org, getPat("GITLAB", org))
		if err != nil {
			log.Error().Err(err).Str("org", org.Name).Msg("failed to connect to GitLab")
			continue
		}
		log.Info().Str("org", org.Name).Str("type", org.Type).Msg("connected to GitLab")
		const perPage = 100
		var all_projects []*gitlab.Project
		for page := 1; ; page++ {
			opt := &gitlab.ListProjectsOptions{
				LastActivityAfter: gitlab.Time(time_ago),
				Membership:        gitlab.Bool(true),
				Simple:            gitlab.Bool(true),
				OrderBy:           gitlab.String("created_at"),
				Sort:              gitlab.String("desc"),
				ListOptions: gitlab.ListOptions{
					PerPage: perPage,
					Page:    page,
				},
			}
			projects, resp, err := git.Projects.ListProjects(opt)
			all_projects = append(all_projects, projects...)
			if err != nil {
				log.Error().Err(err).Str("org", org.Name).Msg("failed to get GitLab projects")
				break
			}
			if resp.CurrentPage >= resp.TotalPages {
				break
			}
		}
		for _, project := range all_projects {
			if !contains(conf.SkipRepos, project.PathWithNamespace) {
				gitlab_repos[project.WebURL] = gitlab_to_git(project, getPat("GITLAB", org), org.Name)
			}
		}
	}
	return gitlab_repos
}
