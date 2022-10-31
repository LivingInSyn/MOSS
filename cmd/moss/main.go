package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
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

func scan_repo(repo *github.Repository, pat, orgname, gl_conf_path string, additional_args []string, results chan GitleaksRepoResult) {
	// build a result object
	result := GitleaksRepoResult{
		Repository: *repo.Name,
		URL:        *repo.HTMLURL,
		IsPrivate:  *repo.Private,
		Org:        orgname,
	}
	// make temp dir
	dir, err := os.MkdirTemp(os.TempDir(), "moss_")
	if err != nil {
		log.Error().Err(err).Str("repo", *repo.Name).Msg("failed to create temp dir to scan repo")
		result.Err = err
		results <- result
		return
	}
	log.Debug().Str("repo", *repo.Name).Str("dir", dir).Msg("tempdir set")
	defer os.RemoveAll(dir)
	// clone into it
	cloneUrl := *repo.CloneURL
	cloneUrl = strings.Replace(cloneUrl, "https://", fmt.Sprintf("https://%s@", pat), 1)
	cloneargs := []string{"clone", cloneUrl, dir}
	cmd := exec.Command("git", cloneargs...)
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("repo", *repo.Name).Msg("failed to clone repo")
		result.Err = err
		results <- result
		return
	}
	// run gitleaks
	outputpath := fmt.Sprintf("%s/__gitleaks.json", dir)
	outputarg := fmt.Sprintf("-r=%s", outputpath)
	confpath := fmt.Sprintf("-c=%s", gl_conf_path)
	// not exactly sure why gitleaks doesn't detect that
	// it IS a git repo, but we can still detect secrets
	dirarg := fmt.Sprintf("-s=%s", dir)
	gitleaks_args := []string{"detect", "-v", "-f=json", "--exit-code=0", outputarg, confpath, dirarg}
	gitleaks_args = append(gitleaks_args, additional_args...)
	// TEMP
	var outb, errb bytes.Buffer
	gl_cmd := exec.Command("gitleaks", gitleaks_args...)
	gl_cmd.Stdout = &outb
	gl_cmd.Stderr = &errb
	//fmt.Println(strings.Join(gl_cmd.Args, " "))
	log.Debug().Str("repo", *repo.FullName).Msg("starting gitleaks scan")
	if err := gl_cmd.Run(); err != nil {
		log.Error().Err(err).Str("repo", *repo.Name).Msg("error running gitleaks on the repo")
		result.Err = err
		results <- result
		return
	}
	log.Debug().Str("repo", *repo.FullName).Msg("finished gitleaks scan")

	// code useful for debugging, but not for leaving compiled
	// fmt.Println(outb.String())
	// fmt.Println(errb.String())
	// log.Debug().Str("stdout", outb.String()).Str("stderr", errb.String()).Msg("output from gitleaks")

	// load the result into a GitleaksResult
	resultfile, err := os.ReadFile(outputpath)
	if err != nil {
		log.Error().Err(err).Str("repo", *repo.Name).Msg("error opening results file")
		result.Err = err
		results <- result
		return
	}
	jsonResults := make([]GitleaksResult, 0)
	err = json.Unmarshal(resultfile, &jsonResults)
	if err != nil {
		log.Error().Err(err).Str("repo", *repo.Name).Msg("error unmarshaling gitleaks results")
		result.Err = err
		results <- result
		return
	}
	//success: return
	result.Results = jsonResults
	result.Err = nil
	results <- result
}

func skip_repo(repo *github.Repository, skipRepos []string) bool {
	for _, s := range skipRepos {
		if s == *repo.FullName {
			return true
		}
	}
	return false
}

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

func main() {
	// setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("logging setup")
	if os.Getenv("MOSS_DEBUG") != "" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("loglevel debug")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	// load the config file
	foo := os.Environ()
	_ = foo
	confdir := os.Getenv("MOSS_CONFDIR")
	if confdir == "" {
		confdir = "./configs/conf.yml"
	}
	var conf Conf
	conf.getConfig(confdir)
	// check the gitleaks.toml file exists and isn't empty
	gitleaks_toml_path := os.Getenv("MOSS_GITLEAKSCONF")
	if gitleaks_toml_path == "" {
		log.Debug().Msg("gitleaks toml path was empty, using default")
		gitleaks_toml_path = "./configs/gitleaks.toml"
	}
	check_gitleaks_conf(gitleaks_toml_path)
	// check the PAT exists for each org
	pats := make(map[string]string, 0)
	for _, org := range conf.GithubConfig.OrgsToScan {
		patenv := fmt.Sprintf("PAT_%s", org)
		pat := os.Getenv(patenv)
		if pat == "" {
			log.Error().Str("org", org).Msg("PAT for org doesn't exist. Skipping it")
			continue
		}
		pats[org] = pat
	}
	if len(pats) == 0 {
		log.Fatal().Msg("No GitHub PATs found, nothing to scan!")
	}
	// foreach org, get the repos according to days_to_scan
	all_repos := make(map[string]*github.Repository, 0)
	for org, pat := range pats {
		repos, err := get_org_repos(org, pat, conf.GithubConfig.DaysToScan, conf.SkipRepos)
		if err != nil {
			log.Error().Err(err).Str("org", org).Msg("Failed to get repos from org. Continuing")
			continue
		}
		for _, repo := range repos {
			all_repos[*repo.HTMLURL] = repo
		}
	}
	// add a useful debug feature here for large github orgs
	repo_limit_s := os.Getenv("MOSS_DEBUG_LIMIT")
	if repo_limit_s != "" {
		repo_limit, err := strconv.Atoi(repo_limit_s)
		if err != nil {
			log.Error().Err(err).Str("MOSS_DEBUG_LIMIT", repo_limit_s).
				Msg("failed to cast value for moss debug limit, setting to 10")
			repo_limit = 10
		}
		limit_repos := make(map[string]*github.Repository, 0)
		counter := 0
		for _, repo := range all_repos {
			if counter == repo_limit {
				break
			}
			limit_repos[*repo.HTMLURL] = repo
			counter = counter + 1
		}
		all_repos = limit_repos
	}
	// make sure we have repos to scan and blow up if we don't
	if len(all_repos) == 0 {
		log.Fatal().Msg("no repos found to scan!")
	}

	// create the channel and kick off the scans
	results := make(chan GitleaksRepoResult, runtime.NumCPU())
	for _, repo := range all_repos {
		reponame := repo.GetFullName()
		orgname := strings.Split(reponame, "/")[0]
		pat := pats[orgname]
		go scan_repo(repo, pat, orgname, gitleaks_toml_path, conf.GitLeaksConfig.AdditionalArgs, results)
	}
	// collect the results
	collected := 0
	final_results := make([]GitleaksRepoResult, 0)
	for {
		repoResult := <-results
		repoResult.filterResults(conf)
		final_results = append(final_results, repoResult)
		collected = collected + 1
		log.Debug().Float32("percent_done", float32(collected)/float32(len(all_repos))).Msg("percent done")
		if collected >= len(all_repos) {
			break
		}
	}
	// format and output the results nicely
	output_dir := os.Getenv("MOSS_OUTDIR")
	if output_dir == "" {
		output_dir = "/output"
	}
	if strings.ToLower(conf.Output.Format) == "json" {
		output := json_output(final_results, conf.GithubConfig.OrgsToScan)
		// todo: make this part of the conf
		outpath := fmt.Sprintf("%s/output.json", output_dir)
		os.WriteFile(outpath, []byte(output), 0644)
	} else if strings.ToLower(conf.Output.Format) == "html" {
		err := html_output(final_results, conf.GithubConfig.OrgsToScan, "")
		if err != nil {
			log.Error().Err(err).Msg("Error creating html output")
		}
	} else if strings.ToLower(conf.Output.Format) == "markdown" {
		mdown_out := markdown_output(final_results, conf.GithubConfig.OrgsToScan)
		outpath := fmt.Sprintf("%s/output.md", output_dir)
		os.WriteFile(outpath, []byte(mdown_out), 0644)
	}
}
