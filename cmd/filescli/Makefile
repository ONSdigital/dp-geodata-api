# stolen from https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-z0-9A-Z_.-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: filescli
filescli:	## build the cli binary
	go build -o filescli ./main.go

.PHONY: mandoc-image
mandoc-image:	## create docker image with mandoc
	docker build -f Dockerfile-mandoc -t local/mandoc:latest .

filescli.md: filescli.mandoc	## generate man page in markdown
	./mandoc.sh -T markdown filescli.mandoc > filescli.md.new && mv filescli.md.new filescli.md

filescli.pdf: filescli.mandoc	## generate pdf man pagae
	./mandoc.sh -T pdf filescli.mandoc > filescli.pdf.new && mv filescli.pdf.new filescli.pdf
