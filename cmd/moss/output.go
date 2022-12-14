package main

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

func get_json_obj(results []GitleaksRepoResult, orgs []string) map[string][]GitleaksRepoResult {
	json_res := make(map[string][]GitleaksRepoResult)
	for _, org := range orgs {
		org_results := make([]GitleaksRepoResult, 0)
		for _, res := range results {
			if res.Org == org {
				org_results = append(org_results, res)
			}
		}
		json_res[org] = org_results
	}
	return json_res
}

func json_output(results []GitleaksRepoResult, orgs []string) string {
	json_res := get_json_obj(results, orgs)
	j_string, err := json.Marshal(json_res)
	if err != nil {
		log.Fatal().Msg("Failed to marshal json results")
	}
	return string(j_string)
}

func markdown_output(results []GitleaksRepoResult, orgs []string) string {
	json_res := get_json_obj(results, orgs)
	markdown_out := "# MOSS Results\n"
	for org, results := range json_res {
		markdown_out = fmt.Sprintf("%s## %s\n", markdown_out, org)
		// check if there were no results
		if len(results) == 0 {
			log.Info().Str("org", org).Msg("no results in org")
			markdown_out = fmt.Sprintf("%sNo findings!\n", markdown_out)
			continue
		}
		// check if the org is clean
		cleanOrg := true
		for _, repo_result := range results {
			if len(repo_result.Results) > 0 {
				cleanOrg = false
				break
			}
		}
		if cleanOrg {
			log.Info().Str("org", org).Msg("org had results, but no detected secrets")
			markdown_out = fmt.Sprintf("%sNo findings!\n", markdown_out)
			continue
		}
		// otherwise, process the results
		// foreach result in results, put a row in the table
		for _, repo_result := range results {
			if len(repo_result.Results) == 0 {
				log.Debug().Str("repo", repo_result.Repository).Msg("skipping due to no findings")
				continue
			}
			// repo header
			markdown_out = fmt.Sprintf("%s### %s\n", markdown_out, repo_result.Repository)
			// start a table
			markdown_out = fmt.Sprintf("%s<details>\n  <summary>Repository Details</summary>\n\n", markdown_out)
			markdown_out = fmt.Sprintf("%s|File Link|Type|Secret|Commit|\n|---------|----|------|------|\n", markdown_out)
			// foreach finding, add a row
			for _, finding := range repo_result.Results {
				row := "|"
				// file
				row = fmt.Sprintf("%s%s|", row, finding.File)
				// type
				row = fmt.Sprintf("%s%s|", row, finding.Description)
				// secret
				if len(finding.Secret) > 10 {
					row = fmt.Sprintf("%s%s|", row, finding.Secret[:10])
				} else {
					row = fmt.Sprintf("%s%s|", row, finding.Secret)
				}
				// commit
				commit_url := fmt.Sprintf("%s/commit/%s", repo_result.URL, finding.Commit)
				commit_link := fmt.Sprintf("[%s](%s)", finding.Commit[:7], commit_url)
				row = fmt.Sprintf("%s%s|", row, commit_link)
				// append the rown to markdown_out and add a newline
				markdown_out = fmt.Sprintf("%s%s\n", markdown_out, row)
			}
			//add an additional newline to make the markdown nice
			markdown_out = fmt.Sprintf("%s\n</details>\n\n", markdown_out)
		}
	}
	return markdown_out
}

func html_output(results []GitleaksRepoResult, orgs []string, outpath string) error {
	json_res := get_json_obj(results, orgs)
	_ = json_res
	return nil
}
