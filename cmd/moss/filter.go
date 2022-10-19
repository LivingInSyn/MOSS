package main

func (r *GitleaksRepoResult) filterResults(conf Conf) {
	// filter out explicitly ignored secrets
	no_ignored := make([]GitleaksResult, 0)
	for _, r := range r.Results {
		ignored := false
		for _, to_ignore := range conf.IgnoreSecrets {
			if to_ignore == r.Secret {
				ignored = true
				break
			}
		}
		if !ignored {
			no_ignored = append(no_ignored, r)
		}
	}
	r.Results = no_ignored

	// filter out  by pattern
	// TODO: implement

	// filter out files for particular repos by regex
	// now filter using those regexes
	no_ignored = make([]GitleaksResult, 0)
	for _, result := range r.Results {
		ignored := false
		for _, expr := range conf.r_ignore_map[r.Repository] {
			if expr.Match([]byte(result.File)) {
				ignored = true
				break
			}
		}
		if !ignored {
			no_ignored = append(no_ignored, result)
		}
	}
	r.Results = no_ignored

}
