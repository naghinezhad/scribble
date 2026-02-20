#@IgnoreInspection BashAddShebang
ROOT=$(realpath $(dir $(lastword $(MAKEFILE_LIST))))

APP_NAME=scribble

GO_CMD?=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=0

# Install by `go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<SET VERSION>`
GOLANGCI_LINT_CMD=$(GO_CMD) tool golangci-lint

NPM_CMD?=npm

.DEFAULT_GOAL := .default

.default: format lint build test

.PHONY: help
help: ## Show help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dep
dep: go-mod-download npm-install ## Install all dependencies

.PHONY: run
run: npm-build go-run ## Run the application after building JS/CSS

.PHONY: format
format: go-format npm-format ## Format all files

.PHONY: lint
lint: npm-lint go-lint ## Lint all files

.PHONY: build
build: npm-build go-build ## Build all components

.PHONY: test
test: go-test ## Run all tests

### Go

.which-go:
	@which $(GO_CMD) > /dev/null || (echo "Install Go from https://go.dev/doc/install" & exit 1)

.PHONY: go-mod-download
go-mod-download: .which-go ## Install go dependencies
	$(GO_CMD) mod download

.PHONY: go-run
go-run: .which-go ## Run the application
	$(GO_CMD) run $(ROOT)/cmd/$(APP_NAME)

.PHONY: go-format
go-format: .which-go ## Format files
	$(GO_CMD) mod tidy
	$(GO_CMD) fix $(ROOT)/...
	$(GOLANGCI_LINT_CMD) fmt $(ROOT)/...

.PHONY: go-lint
go-lint: .which-go ## Lint files
	$(GOLANGCI_LINT_CMD) run $(ROOT)/...

.PHONY: go-test
go-test: .which-go ## Run tests
	CGO_ENABLED=$(CGO_ENABLED) $(GO_CMD) test -race -cover -coverprofile=coverage.out -covermode=atomic $(ROOT)/...

.PHONY: go-build
go-build: .which-go ## Build binary
	$(GO_CMD) build -v -trimpath -ldflags="-s -w" -o $(ROOT)/bin/$(APP_NAME)_$(GOOS)_$(GOARCH) $(ROOT)/cmd/$(APP_NAME)

### Node

.which-npm:
	@which $(NPM_CMD) > /dev/null || (echo "Install NodeJS from https://nodejs.org/en/download" & exit 1)

.PHONY: npm-install
npm-install: .which-npm ## Install JS dependencies
	$(NPM_CMD) --prefix $(ROOT)/web install

.PHONY: npm-build
npm-build: .which-npm ## Build JS and CSS
	$(NPM_CMD) --prefix $(ROOT)/web run build:js
	$(NPM_CMD) --prefix $(ROOT)/web run build:css

.PHONY: npm-format
npm-format: .which-npm ## Format JS and CSS files
	$(NPM_CMD) --prefix $(ROOT)/web run format

.PHONY: npm-lint
npm-lint: .which-npm ## Lint JS and CSS files
	$(NPM_CMD) --prefix $(ROOT)/web run lint
