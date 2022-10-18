package main

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

func json_output(results []GitleaksRepoResult, orgs []string) string {
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
	j_string, err := json.Marshal(json_res)
	if err != nil {
		log.Fatal().Msg("Failed to marshal json results")
	}
	return string(j_string)
}
