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
	if len(conf.GithubConfig.OrgsToScan) != 2 {
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
