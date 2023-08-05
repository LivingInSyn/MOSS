package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/go-github/v47/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/semaphore"
)

func check_gitleaks_conf(gitleaks_path string) error {
	_, err := os.ReadFile(gitleaks_path)
	if err != nil {
		log.Fatal().Err(err).Str("path", gitleaks_path).Msg("failed to read gitleaks toml")
		return err
	}
	return nil
}

func scan_repo(repo *GitRepo, gl_conf_path string, additional_args []string, results chan GitleaksRepoResult, sem *semaphore.Weighted) {
	//Semaphone logic for Max Concurrencies
	ctx := context.Background()
	if err := sem.Acquire(ctx, 1); err != nil {
		// log.Printf("Failed to acquire semaphore: %v", err)
		log.Fatal().Err(err).Msg("failed to lock a semaphore")
	}
	defer sem.Release(1)
	// build a result object
	result := GitleaksRepoResult{
		Repository: repo.Name,
		URL:        repo.HTMLURL,
		IsPrivate:  repo.Private,
		Org:        repo.orgname,
	}
	// make temp dir
	dir, err := os.MkdirTemp(os.TempDir(), "moss_")
	if err != nil {
		log.Error().Err(err).Str("repo", repo.Name).Msg("failed to create temp dir to scan repo")
		result.Err = err
		results <- result
		return
	}
	log.Debug().Str("repo", repo.Name).Str("dir", dir).Msg("tempdir set")
	defer os.RemoveAll(dir)
	// clone into it
	cloneUrl := repo.CloneURL
	if repo.provider == "GITHUB" {
		cloneUrl = strings.Replace(cloneUrl, "https://", fmt.Sprintf("https://%s@", repo.pat), 1)
	} else if repo.provider == "GITLAB" {
		cloneUrl = strings.Replace(cloneUrl, "https://", fmt.Sprintf("https://oauth2:%s@", repo.pat), 1)
	}
	cloneargs := []string{"clone", cloneUrl, dir}
	cmd := exec.Command("git", cloneargs...)
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("repo", repo.Name).Msg("failed to clone repo")
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
	log.Debug().Str("repo", repo.FullName).Msg("starting gitleaks scan")
	if err := gl_cmd.Run(); err != nil {
		log.Error().Err(err).Str("repo", repo.Name).Msg("error running gitleaks on the repo")
		result.Err = err
		results <- result
		return
	}
	log.Debug().Str("repo", repo.FullName).Msg("finished gitleaks scan")

	// code useful for debugging, but not for leaving compiled
	// fmt.Println(outb.String())
	// fmt.Println(errb.String())
	// log.Debug().Str("stdout", outb.String()).Str("stderr", errb.String()).Msg("output from gitleaks")

	// load the result into a GitleaksResult
	resultfile, err := os.ReadFile(outputpath)
	if err != nil {
		log.Error().Err(err).Str("repo", repo.Name).Msg("error opening results file")
		result.Err = err
		results <- result
		return
	}
	jsonResults := make([]GitleaksResult, 0)
	err = json.Unmarshal(resultfile, &jsonResults)
	if err != nil {
		log.Error().Err(err).Str("repo", repo.Name).Msg("error unmarshaling gitleaks results")
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

// Fetch the PATs for the respective ORGs
func getPats(provider string, orgs []string) map[string]string {
	var orgPats = map[string]string{}
	for _, org := range orgs {
		log.Debug().Str("env_var", provider+"_PAT_"+org).Msg("trying to get env var")
		pat := os.Getenv(provider + "_PAT_" + org)
		if pat == "" {
			log.Error().Str("org", org).Str("provider", provider).Msg("provider PAT for org doesn't exist. Skipping it")
			continue

		}
		orgPats[org] = pat
	}
	return orgPats
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
	// Fetch the PATs for respective Provider
	github_pats := getPats("GITHUB", conf.GithubConfig.OrgsToScan)
	gitlab_pats := getPats("GITLAB", conf.GitlabConfig.OrgsToScan)
	//Check for scanning single repository
	repoURL := flag.String("repo", "", "Repository URL to scan")
	flag.Parse()
	//collate all the repos
	all_repos := make(map[string]*GitRepo, 0)
	for key, value := range get_all_github_repos(github_pats, conf) {
		all_repos[key] = value
	}
	gitlab_repos, nil := get_all_gitlab_repos(gitlab_pats, conf.GitlabConfig.DaysToScan, conf.SkipRepos)
	for key, value := range gitlab_repos {
		all_repos[key] = value
	}
	// if we're debugging,  set a limit
	repo_limit_s := os.Getenv("MOSS_DEBUG_LIMIT")
	if repo_limit_s != "" {
		repo_limit, err := strconv.Atoi(repo_limit_s)
		if err != nil {
			log.Error().Err(err).Str("MOSS_DEBUG_LIMIT", repo_limit_s).
				Msg("failed to cast value for moss debug limit, setting to 10")
			repo_limit = 10
		}
		limit_repos := make(map[string]*GitRepo, 0)
		counter := 0
		for _, repo := range all_repos {
			if counter == repo_limit {
				break
			}
			limit_repos[repo.HTMLURL] = repo
			counter = counter + 1
		}
		all_repos = limit_repos
	}
	// make sure we have repos to scan and blow up if we don't
	if len(all_repos) == 0 {
		log.Fatal().Msg("no repos found to scan!")
	}
	// build a semaphor for MaxConcurrency
	sem := semaphore.NewWeighted(conf.MaxConcurrency)
	// create the channel and kick off the scans
	results := make(chan GitleaksRepoResult, runtime.NumCPU())
	if *repoURL != "" {
		log.Debug().Msg("Respository is specified. scanning only the repository")
		// Find the specific repository in the all_repos map
		repo := all_repos[*repoURL]
		if *repo != (GitRepo{}) {
			// Scan the specific repository using the scan_repo function
			scan_repo(repo, gitleaks_toml_path, conf.GitLeaksConfig.AdditionalArgs, results, sem)
			//Clearing the all_repos to make sure the scan is 100%
			all_repos = map[string]*GitRepo{
				*repoURL: all_repos[*repoURL],
			}
		} else {
			log.Fatal().Str("repoURL", *repoURL).Msg("Repository not found in the org")
		}
	} else {
		for _, repo := range all_repos {
			go scan_repo(repo, gitleaks_toml_path, conf.GitLeaksConfig.AdditionalArgs, results, sem)
		}
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
	all_orgs := append(conf.GithubConfig.OrgsToScan, conf.GitlabConfig.OrgsToScan...)
	if strings.ToLower(conf.Output.Format) == "json" {
		output := json_output(final_results, all_orgs)
		// todo: make this part of the conf
		outpath := fmt.Sprintf("%s/output.json", output_dir)
		os.WriteFile(outpath, []byte(output), 0644)
	} else if strings.ToLower(conf.Output.Format) == "html" {
		err := html_output(final_results, all_orgs, "")
		if err != nil {
			log.Error().Err(err).Msg("Error creating html output")
		}
	} else if strings.ToLower(conf.Output.Format) == "markdown" {
		mdown_out := markdown_output(final_results, all_orgs)
		outpath := fmt.Sprintf("%s/output.md", output_dir)
		os.WriteFile(outpath, []byte(mdown_out), 0644)
	}
}