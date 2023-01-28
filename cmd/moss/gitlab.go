package main

import (
	"fmt"

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

// Scan and fetch all the gilab repo list
func get_all_gitlab_repos(gitlabPats map[string]string, daysAgo int, skipRepos []string) (map[string]*GitRepo, error) {
	gitlab_repos := make(map[string]*GitRepo)
	for org, pat := range gitlabPats {
		const perPage = 100
		var all_projects []*gitlab.Project
		git, err := gitlab.NewClient(pat, gitlab.WithBaseURL("https://gitlab."+org+".com"))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to GitLab for org %s: %w", org, err)
		}
		for page := 1; ; page++ {

			opt := &gitlab.ListProjectsOptions{
				Membership: gitlab.Bool(true),
				Simple:     gitlab.Bool(true),
				OrderBy:    gitlab.String("created_at"),
				Sort:       gitlab.String("desc"),
				ListOptions: gitlab.ListOptions{
					PerPage: perPage,
					Page:    page,
				},
			}
			// List all projects for the org
			projects, resp, err := git.Projects.ListProjects(opt)
			all_projects = append(all_projects, projects...)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitLab projects for org %s: %w", org, err)
			}
			if resp.CurrentPage >= resp.TotalPages {
				break
			}
		}

		for _, project := range all_projects {
			if !contains(skipRepos, project.PathWithNamespace) {
				gitlab_repos[project.WebURL] = gitlab_to_git(project)
				gitlab_repos[project.WebURL].pat = pat
				gitlab_repos[project.WebURL].provider = "GITLAB"
				gitlab_repos[project.WebURL].orgname = org
			}
		}
	}
	return gitlab_repos, nil
}
