# MOSS
_a rolling secret gathers no MOSS_ - @robbkidd

MOSS is the Multi-Organization Secret Scanner. It is designed to handle scanning many repositories from multiple github orgs as efficiently as possible. Scanning for secrets is done with [Gitleaks](https://github.com/zricethezav/gitleaks)

## Setting Access Tokens
Organization access tokens (PATs) are passed as env vars in the format `PAT_<orgname>`. So if your orgname is `foo` you would pass the PAT for the account running the scan as: `PAT_foo`. 

MOSS looks for these PATs based on the organizations configured in the `github_config.orgs_to_scan` section of the config file documented below.

## MOSS Config File
A sample configuration file with annotations is [here](./configs/conf.yml)

### Output
The currently supported formats are `markdown` and `json`. Markdown files are written by default to `/output/output.md` but the path where `output.md` can be written to can be set using an environmental variable.

Json files are named `output.json` by default and will also be written to the `/output` folder unless overridden.

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
    -e PAT_someorg=$(GH_TOKEN) \
    -v `pwd`/configs/conf.yml:/usr/src/moss/configs/conf.yml \
    -v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
    -v `pwd`/sample_output:/output \
    --name moss_r \
    ghcr.io/livinginsyn/moss:latest
```

