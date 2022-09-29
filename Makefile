# stolen from https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-z0-9A-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# DEV_HOST and INT_HOST are the EC2 instances running the API for backend testing
# and for front-end development.
# We have to change these hostnames whenever an instance is rebooted because the
# names are based on non-static IPs.
DEV_HOST=ec2-18-193-6-194.eu-central-1.compute.amazonaws.com
INT_HOST=ec2-35-158-105-228.eu-central-1.compute.amazonaws.com
SSH_KEY=swaggerui/frank-ec2-dev0.pem

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
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-geodata-api ./cmd/dp-geodata-api

.PHONY: build-linux-amd
build-linux-amd:	## build poc service specifically for linux on amd64 (used for EC2 deploy)
	GOOS=linux GOARCH=amd64 go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-geodata-api.amd ./cmd/dp-geodata-api

.PHONY: build-linux-arm
build-linux-arm:	## build poc service specifically for linux on arm64 (used for EC2 deploy)
	GOOS=linux GOARCH=arm64 go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-geodata-api.arm ./cmd/dp-geodata-api

.PHONY: debug
debug:	## run poc service in debug mode
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-geodata-api ./cmd/dp-geodata-api
	HUMAN_LOG=1 DEBUG=1 DO_CORS=true $(BINPATH)/dp-geodata-api

.PHONY: test
test:	## run poc tests
	go test -race -cover ./...
	cd data-tiles && make test

.PHONY: cover
cover:	## aggregate coverage hat tip @efragkiadaki
	go test -count=1 -coverprofile=coverage.out ./...
	@awk 'BEGIN {cov=0; stat=0;} $$3!="" { cov+=($$3==1?$$2:0); stat+=$$2; } END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}' coverage.out
	@go tool cover -html=coverage.out

.PHONY: test-component
test-component:	## blank target (for now) to satisfy CI
	exit

.PHONY: test-integration
test-integration:	## integration tests needs web server
	go test -count=1 ./inttests -tags=integration

.PHONY: test-datasanity
test-datasanity:	## this needs a DB and postgres env vars set
	go test -count=1 ./dataingest/datasanity  -tags=datasanity

.PHONY: test-comptest
test-comptest:	## this provisions a docker DB automatically
	go test -p 1 -count=1 ./...  -tags=comptest

.PHONY: test-comptestv
test-comptestv:	## this provisions a docker DB automatically
	go test -v -p 1 -count=1 ./...  -tags=comptest

.PHONY: test-comptest-kill
test-comptest-kill:	## this provisions a docker DB automatically & kills it
	go test -p 1 -count=1 ./...  -tags=comptest -args -kill=true

.PHONY: convey
convey:	## run goconvey
	goconvey ./...

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
generate:	## update autogenerated files
	@go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.8.2
	@go install github.com/pseudo-su/oapi-ui-codegen/cmd/oapi-ui-codegen@v0.0.2

	@go generate cmd/dp-geodata-api/generate.go

#
# check no manual commits to auto-generated files on a clean tree
#

.PHONY: check-generate
check-generate: generate	## check no manual commits to auto-generated files on a clean tree
	  @[ "$$(git diff)" = "" ] || (echo "commits to autogen file?" && exit 1)


# local docker
#
.PHONY: image
image:	## create docker image for local api
	docker build -t dp-geodata-api -f Dockerfile.api .

.PHONY: update-schema-image
update-schema-image:	## build the update-schema image
	docker build -t update-schema -f Dockerfile.update-schema .

.PHONY: run-update-schema
run-update-schema:	## run update-schema in a local container
	docker run \
		-it \
		--env PGHOST="$${PGHOST_INTERNAL:-$$PGHOST}" \
		--env PGPORT="$${PGPORT_INTERNAL:-$$PGPORT}" \
		--env PGDATABASE \
		--env PGUSER \
		--env PGPASSWORD \
		--env POSTGRES_PASSWORD \
		--rm \
		update-schema

#
# ssh to EC2 instances
#
ssh-dev:	## ssh to dev EC2 instance
	chmod 0600 $(SSH_KEY)
	ssh -i $(SSH_KEY) ubuntu@$(DEV_HOST)

ssh-int:	## ssh to integration EC2 instance
	chmod 0600 $(SSH_KEY)
	ssh -i $(SSH_KEY) ubuntu@$(INT_HOST)

#
# deploy to EC2 instances
#
deploy-dev:	## deploy build/dp-geodata-api.amd to dev EC2 instance
	scp -i $(SSH_KEY) build/dp-geodata-api.amd ubuntu@$(DEV_HOST):dp-geodata-api.new
	ssh -i $(SSH_KEY) ubuntu@$(DEV_HOST) ./deploy.sh dp-geodata-api.new

deploy-int:	## deploy build/dp-geodata-api.arm to F/E EC2 instance
	scp -i $(SSH_KEY) build/dp-geodata-api.arm ubuntu@$(INT_HOST):dp-geodata-api.new
	ssh -i $(SSH_KEY) ubuntu@$(INT_HOST) ./deploy.sh dp-geodata-api.new

# integration tests
#
# If you want to run anything other than "make test" in inttests/,
# then you should set $TEST_TARGET_URL in your environment, and then
# run make directly from within inttests/.
# These targets are just shortcuts for post-deploy sanity tests.
test-local: ## run integration tests against local instance
	cd inttests && make test-local

test-dev:	## run integration tests against dev EC2 instance
	cd inttests && TEST_TARGET_URL=http://$(DEV_HOST):25252 DO_CORS=true make test

test-int:	## run integration tests against F/E EC2 instance
	cd inttests && TEST_TARGET_URL=http://$(INT_HOST):25252 DO_CORS=true make test

#
# rollback API on EC2 instances
#
rollback-dev:	## rollback API on dev EC2 instance
	ssh -i $(SSH_KEY) ubuntu@$(DEV_HOST) ./deploy.sh previous

rollback-int:	## rollback API on F/E EC2 instance
	ssh -i $(SSH_KEY) ubuntu@$(INT_HOST) ./deploy.sh previous

#
# get healthcheck from EC2 instances
#
health-dev:	## get healthcheck from dev EC2 instance
	curl -s http://$(DEV_HOST):25252/health | jq

health-int:	## get healthcheck from F/E EC2 instance
	curl -s http://$(INT_HOST):25252/health | jq
