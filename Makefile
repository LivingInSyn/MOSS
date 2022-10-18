# note: call scripts from /scripts

build:
	docker build -t moss .
run:
	docker run -it --rm \
		-e PAT_puppetlabs=$(GH_TOKEN) \
		-e MOSS_DEBUG=true \
		-v `pwd`/conf_test.yml:/usr/src/moss/configs/conf.yml \
		-v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
		--name moss_r \
		moss

run_debug:
	docker run -it --rm \
		-e PAT_puppetlabs=$(GH_TOKEN) \
		-e MOSS_DEBUG=true \
		-v `pwd`/conf_test.yml:/usr/src/moss/configs/conf.yml \
		-v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
		--name moss_r \
		moss /bin/bash