package main

import "testing"

func getConf() Conf {
	ghc := ConfGithubConfig{
		OrgsToScan: []string{"org"},
		DaysToScan: 3,
	}
	c := Conf{
		GithubConfig:         ghc,
		SkipRepos:            make([]string, 0),
		IgnoreSecretPatterns: make([]string, 0),
		IgnoreSecrets:        make([]string, 0),
		ReposToIgnore:        map[string][]string{},
		Output:               ConfOutput{Format: "json"},
	}
	c.buildIgnoreMap()
	return c
}

func TestFilterSuppressedSecret(t *testing.T) {
	// setup the secret result
	rr := GitleaksRepoResult{
		Repository: "repo",
		Org:        "org",
		URL:        "https://github.com/org/repo",
		Err:        nil,
		IsPrivate:  true,
		Results:    make([]GitleaksResult, 0),
	}
	r := GitleaksResult{
		Description: "Generic API Key",
		StartLine:   63,
		EndLine:     64,
		StartColumn: 16,
		EndColumn:   1,
		Match:       "Keygrip = DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		Secret:      "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		File:        "somefolder/README.md",
		Commit:      "BEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEAD",
		Entropy:     3.6257162,
		Author:      "John Smit",
		Email:       "john.smith@example.com",
		Date:        "2021-07-02T18:44:24Z",
		Message:     "(maint) Clarify gpg-preset-passphrase instructions.",
		Tags:        make([]interface{}, 0),
		RuleID:      "generic-api-key",
	}
	rr.Results = append(rr.Results, r)
	// setup a conf
	c := getConf()
	// set the secret to ignore
	c.IgnoreSecrets = append(c.IgnoreSecrets, "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF")
	// run the filter and make sure there are 0 results now
	rr.filterResults(c)
	if len(rr.Results) > 0 {
		t.Errorf("secret wasn't filtered out!")
	}

}

func TestFilterSuppressedPattern(t *testing.T) {
	// setup the secret result
	rr := GitleaksRepoResult{
		Repository: "repo",
		Org:        "org",
		URL:        "https://github.com/org/repo",
		Err:        nil,
		IsPrivate:  true,
		Results:    make([]GitleaksResult, 0),
	}
	r := GitleaksResult{
		Description: "Generic API Key",
		StartLine:   63,
		EndLine:     64,
		StartColumn: 16,
		EndColumn:   1,
		Match:       "Keygrip = DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		Secret:      "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		File:        "somefolder/README.md",
		Commit:      "BEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEAD",
		Entropy:     3.6257162,
		Author:      "John Smit",
		Email:       "john.smith@example.com",
		Date:        "2021-07-02T18:44:24Z",
		Message:     "(maint) Clarify gpg-preset-passphrase instructions.",
		Tags:        make([]interface{}, 0),
		RuleID:      "generic-api-key",
	}
	rr.Results = append(rr.Results, r)
	// setup a conf
	c := getConf()
	// set the secret to ignore
	c.ReposToIgnore["org/repo"] = []string{"somefolder/.*"}
	c.buildIgnoreMap()
	// run the filter and make sure there are 0 results now
	rr.filterResults(c)
	if len(rr.Results) > 0 {
		t.Errorf("secret wasn't filtered out!")
	}

}

func TestFilterSuppressedCommit(t *testing.T) {
	// setup the secret result
	rr := GitleaksRepoResult{
		Repository: "repo",
		Org:        "org",
		URL:        "https://github.com/org/repo",
		Err:        nil,
		IsPrivate:  true,
		Results:    make([]GitleaksResult, 0),
	}
	r := GitleaksResult{
		Description: "Generic API Key",
		StartLine:   63,
		EndLine:     64,
		StartColumn: 16,
		EndColumn:   1,
		Match:       "Keygrip = DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		Secret:      "DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		File:        "somefolder/README.md",
		Commit:      "BEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEAD",
		Entropy:     3.6257162,
		Author:      "John Smit",
		Email:       "john.smith@example.com",
		Date:        "2021-07-02T18:44:24Z",
		Message:     "(maint) Clarify gpg-preset-passphrase instructions.",
		Tags:        make([]interface{}, 0),
		RuleID:      "generic-api-key",
	}
	rr.Results = append(rr.Results, r)
	// setup a conf
	c := getConf()
	// set the secret to ignore
	c.IgnoreCommits = append(c.IgnoreCommits, "BEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEAD")
	// run the filter and make sure there are 0 results now
	rr.filterResults(c)
	if len(rr.Results) > 0 {
		t.Errorf("secret wasn't filtered out!")
	}

}
