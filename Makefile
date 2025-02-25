# cf. based https://gist.github.com/thomaspoignant/5b72d579bd5f311904d973652180c705
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=xk6-gtp
# brunch name version
VERSION := v0.0.1
DOCKER_REGISTRY?= ghcr.io/bbsakura/
DIFF_FROM_BRANCH_NAME ?= origin/main

ENTRY_POINT_DIR=cmd
TARGETS=$(notdir $(wildcard $(ENTRY_POINT_DIR)/*))

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test stylecheck build make_outdir single_target clean xk6build package help $(TARGETS)

all: help

## Build:
build: make_outdir single_target $(TARGETS) xk6build ## Build your project and put the output binary in out/bin/
make_outdir:
	mkdir -p out/bin

single_target:
	$(GOCMD) build -o out/bin/main ./main.go

$(TARGETS):
	$(GOCMD) build -o out/bin/$@ ./cmd/$@/...

clean: ## Remove build related file
	rm -fr ./out/bin

xk6build: ## Package the project
	xk6 build --with github.com/grafana/xk6-output-prometheus-remote@v0.3.1 \
	--with github.com/bbsakura/xk6-gtp@latest=$(shell pwd) --output ./out/bin/xk6

## Test:
test: ## Run the tests of the project
	$(GOTEST) -v -race ./... $(OUTPUT_OPTIONS)
	./scripts/run-examples.sh

## Lint:
stylecheck: ## Use precommit, fmt and lint for this project.
	pre-commit run --show-diff-on-failure --color=always --all-files

stylecheck-ci: ## Run pre-commit for CI
	pre-commit run --show-diff-on-failure --color=always --from-ref $(DIFF_FROM_BRANCH_NAME) --to-ref HEAD

## Docker:
docker-build: ## Use the dockerfile to build the container
	docker build --rm --tag $(BINARY_NAME) .

docker-release: ## Release the container with tag latest and version
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)
	# Push the docker images
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)

## Golang:
install-go-tools: ## install project go tools
	sh scripts/install-go-tools.sh

go-gen: ## go:generate invocations
	go generate ./...

## mise:
install-dev-pkg: ## install .tool-version
	mise install

# go-dependency-sync
go-dep-sync:
	go install github.com/grafana/go-depsync@latest
	go-depsync --gomod go.mod --parent go.k6.io/k6 2| sh

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_0-9-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

