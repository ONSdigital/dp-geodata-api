# stolen from https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-z0-9A-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
VERSIONDIRTY := $(shell git diff --quiet HEAD; git describe --tags --always --long --dirty | sed 's/-/+/' | sed 's/-/./g')

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
LDFLAGSDIRTY = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSIONDIRTY)"

.PHONY: all	## run audit, test and build
all: audit test build

.PHONY: audit
audit:	## run nancy auditor
	go list -json -m all | nancy sleuth --exclude-vulnerability-file ./.nancy-ignore

.PHONY: lint
lint:	## doesn't really lint
	exit

.PHONY: build
build:	## build poc service
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-find-insights-poc-api ./cmd/dp-find-insights-poc-api

.PHONY: build-linux-amd
build-linux-amd:	## build poc service specifically for linux on amd64 (used for EC2 deploy)
	GOOS=linux GOARCH=amd64 go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-find-insights-poc-api ./cmd/dp-find-insights-poc-api

.PHONY: debug
debug:	## run poc service in debug mode
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-find-insights-poc-api ./cmd/dp-find-insights-poc-api
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-find-insights-poc-api

.PHONY: test
test:	## run poc tests
	go test -race -cover ./...

.PHONY: cover
cover:	## aggregate coverage hat tip @efragkiadaki
	go test -count=1 -coverprofile=coverage.out ./...
	@awk 'BEGIN {cov=0; stat=0;} $$3!="" { cov+=($$3==1?$$2:0); stat+=$$2; } END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}' coverage.out
	@go tool cover -html=coverage.out

.PHONY: test-integration
test-integration:	## integration tests needs web server
	go test -count=1 ./inttests -tags=integration

.PHONY: test-datasanity
test-datasanity:	## this needs a DB and postgres env vars set
	go test -count=1 ./dataingest/datasanity  -tags=datasanity

.PHONY: test-comptest
test-comptest:	## this provisions a docker DB automatically
	go test -count=1 ./...  -tags=comptest

.PHONY: test-comptestv
test-comptestv:	## this provisions a docker DB automatically
	go test -v -count=1 ./...  -tags=comptest

.PHONY: test-comptest-kill
test-comptest-kill:	## this provisions a docker DB automatically & kills it
	go test -count=1 ./...  -tags=comptest -args -kill=true

.PHONY: convey
convey:	## run goconvey
	goconvey ./...

#
# these are the lambda-related targets
#

.PHONY: build-lambda
build-lambda:	## compile lambda
	GOOS=linux GOARCH=amd64 go build -o build/hello ./functions/hello/...

.PHONY: bundle-lambda
bundle-lambda:	## bundle lambda into .zip to deploy
	zip -j build/hello.zip build/hello

#
# cli targets
#

.PHONY: build-cli
build-cli:	## build the hello cli (build/geodata)
	go build -o build/geodata ./cmd/geodata/...

.PHONY: run-cli
run-cli:	## quick sanity test on cli (must set env vars)
	build/geodata --dataset atlas2011.qs119ew

#
# creatschema
#

.PHONY: update-schema
update-schema:
	@go build $(LDFLAGSDIRTY) -o $(BINPATH)/creatschema ./cmd/creatschema/...
	@$(BINPATH)/creatschema

#
# update autogenerated files
# (moq is also used but which version is unclear & manual changes made!)

.PHONY: generate
generate:
	@go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.8.2
	@go install github.com/pseudo-su/oapi-ui-codegen/cmd/oapi-ui-codegen@v0.0.2

	@go generate cmd/dp-find-insights-poc-api/generate.go

#
# check no manual commits to auto-generated files on a clean tree
#

.PHONY: check-generate
check-generate: generate
	  @[ "$$(git diff)" = "" ] || (echo "commits to autogen file?" && exit 1)


# local docker
#
.PHONY: image
image:	## create docker image for local api
	docker build -t dp-find-insights-poc-api -f Dockerfile.api .

.PHONY: run-image
run-api:	## run api in local docker
	docker run \
		-it \
		--rm \
		--publish 127.0.0.1:12550:12550 \
		--env-file secrets/PGPASSWORD.env \
		--name dp-find-insights-poc-api \
		dp-find-insights-poc-api
