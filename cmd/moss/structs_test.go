package main

import (
	"fmt"
	"os"
	"testing"
)

func TestConfigParsing(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory")
	}
	// load it, note that this will explode with a Fatal so we don't
	// really need to catch errors
	confpath := fmt.Sprintf("%s/../../configs/test_conf.yml", cwd)
	var conf Conf
	conf.getConfig(confpath)
	// make sure it parsed correctly
	if len(conf.GithubConfig.OrgsToScan) != 3 {
		t.Errorf("orgs to scan didn't load correctly. Got a len of: %d", len(conf.GithubConfig.OrgsToScan))
	}
	if conf.GithubConfig.DaysToScan != 30 {
		t.Errorf("Got wrong num of days to scan. Wanted 30, got %d", conf.GithubConfig.DaysToScan)
	}
	if len(conf.GitLeaksConfig.AdditionalArgs) != 2 {
		t.Errorf("addtl args didn't load correctly. Got a len of: %d", len(conf.GithubConfig.OrgsToScan))
	}
	if conf.SkipRepos[0] != "some_org/some_repo" {
		t.Errorf("Failed to load skip repos")
	}
	if conf.IgnoreSecrets[0] != "0xDEADBEEF" {
		t.Errorf("failed to parse ignore secrets")
	}
	if conf.IgnoreCommits[0] != "c0a4e7c1208fb49c28b2979fe68985ddac696a6e" {
		t.Errorf("failed to parse ignore commits")
	}
	if conf.ReposToIgnore["some_org/some_repo"][0] != "docs/.*" {
		t.Errorf("failed to get repostoignore")
	}
	if conf.Output.Format != "markdown" {
		t.Errorf("failed to get format correctly")
	}
	var mc int64 = 20
	if conf.MaxConcurrency != mc {
		t.Errorf("got the wrong max concurrency")
	}
}
func TestValidateUniqueOrgNames(t *testing.T) {
	tests := []struct {
		name      string
		conf      Conf
		expectErr bool
	}{
		{
			name: "Valid configuration - no duplicates",
			conf: Conf{
				GithubConfig: ConfGithubConfig{
					OrgsToScan: []OrgConfig{
						{Name: "github-org1"},
						{Name: "github-org2"},
					},
				},
				GitlabConfig: ConfGitlabConfig{
					OrgsToScan: []OrgConfig{
						{Name: "gitlab-org1"},
						{Name: "gitlab-org2"},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Duplicate GitHub organization names",
			conf: Conf{
				GithubConfig: ConfGithubConfig{
					OrgsToScan: []OrgConfig{
						{Name: "duplicate-org"},
						{Name: "duplicate-org"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Duplicate GitLab organization names",
			conf: Conf{
				GitlabConfig: ConfGitlabConfig{
					OrgsToScan: []OrgConfig{
						{Name: "duplicate-org"},
						{Name: "duplicate-org"},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Same name in both GitHub and GitLab is allowed",
			conf: Conf{
				GithubConfig: ConfGithubConfig{
					OrgsToScan: []OrgConfig{
						{Name: "same-name-org"},
					},
				},
				GitlabConfig: ConfGitlabConfig{
					OrgsToScan: []OrgConfig{
						{Name: "same-name-org"},
					},
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.conf.validateUniqueOrgNames()
			if tc.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
