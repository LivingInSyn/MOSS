# MOSS
_a rolling secret gathers no MOSS_ - @robbkidd

MOSS is the Multi-Organization Secret Scanner. It is designed to handle scanning many repositories from multiple github AND gitlab orgs as efficiently as possible. Scanning for secrets is done with [Gitleaks](https://github.com/zricethezav/gitleaks)

## Setting Access Tokens
Organization access tokens (PATs) are passed as env vars.

GitHub tokens will be in the form: `GITHUB_PAT_<orgname>_<CLOUD|ONPREM>`

Gitlab tokens will be in the format `GITLAB_PAT_<orgname>_<CLOUD|ONPREM>`. 

So if you're scanning a github org and the orgname is `foo` you would pass the PAT for the account running the scan as: `GITHUB_PAT_foo_CLOUD` | `GITHUB_PAT_foo_ONPREM`. 

MOSS looks for these PATs based on the organizations configured in the `github_config.orgs_to_scan` section of the config file documented below.

## MOSS Config File
A sample configuration file with annotations is [here](./configs/conf.yml)

## max_concurrency
Care should be taken with max_concurrency. Larger values of max concurrency will result in faster scans* with increased parallelization up to the point of instability. 20 seems to be a reasonable default value. 

## Scanning a specific repository
Specific repositories in an organization can be scanned by adding a flag `repo` to the binary. repo in this case is the HTML URL of the repository. It can be done in the following way
```shell
docker run --rm \
    -e GITLAB_PAT_someorg_CLOUD=$(GL_TOKEN) \
    -v `pwd`/configs/conf.yml:/usr/src/moss/configs/conf.yml \
    -v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
    -v `pwd`/sample_output:/output \
    --name moss_r \
    ghcr.io/livinginsyn/moss:latest -repo=https://gitlab.com/<path_to_repository>
``` 
### Output
The currently supported formats are `markdown` and `json`. Markdown files are written by default to `/output/output.md` but the path where `output.md` can be written to can be set using an environmental variable.

Json files are named `output.json` by default and will also be written to the `/output` folder unless overridden.

Supported formats can be overriden with command-line arguments while running moss 
```shell
moss -format=<json|markdown>
```

## Other Environmental Variables
The following environmental variables may be configured to change the behavior of MOSS:

|Variable Name|Required|Description|Default|
|---|---|---|---|
|MOSS_OUTDIR|False|Sets the directory for MOSS output|/output/{output filename}|
|MOSS_DEBUG|False|Enabled verbose debug logs|False|
|MOSS_CONFDIR|False|Sets the path to the MOSS configuration file|./configs/conf.yml|
|MOSS_GITLEAKSCONF|False|Sets the path to the GitLeaks toml file|./configs/gitleaks.toml|
|MOSS_DEBUG_LIMIT|False|Sets a limit for the number of repos to scan|If not set, it does nothing. If set to an int it is the upper limit, if another string is passed it will default to 10|

## Running with Docker
Docker is the preferred method for running MOSS. A sample run command would be:

```shell
docker run --rm \
    -e GITHUB_PAT_someorg_CLOUD=$(GH_TOKEN) \
    -v `pwd`/configs/conf.yml:/usr/src/moss/configs/conf.yml \
    -v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
    -v `pwd`/sample_output:/output \
    --name moss_r \
    ghcr.io/livinginsyn/moss:latest
```

