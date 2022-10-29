# note: call scripts from /scripts

build:
	docker build -t ghcr.io/livinginsyn/moss:$(MOSS_TAG) .

publish:
	docker push ghcr.io/livinginsyn/moss:$(MOSS_TAG)

run:
	make build
	docker run -it --rm \
		-e PAT_puppetlabs=$(GH_TOKEN) \
		-e MOSS_DEBUG=true \
		-e MOSS_DEBUG_LIMIT=10 \
		-v `pwd`/conf_test.yml:/usr/src/moss/configs/conf.yml \
		-v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
		-v `pwd`/sample_output:/output \
		--name moss_r \
		ghcr.io/livinginsyn/moss:$(MOSS_TAG)

run_debug:
	make build
	docker run -it --rm \
		-e PAT_puppetlabs=$(GH_TOKEN) \
		-e MOSS_DEBUG=true \
		-e MOSS_DEBUG_LIMIT=10 \
		-v `pwd`/conf_test.yml:/usr/src/moss/configs/conf.yml \
		-v `pwd`/configs/gitleaks.toml:/usr/src/moss/configs/gitleaks.toml \
		-v `pwd`/sample_output:/output \
		--name moss_r \
		ghcr.io/livinginsyn/moss:$(MOSS_TAG) /bin/bash