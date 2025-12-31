.PHONY: help build test clean setup install fmt lint docker run release

REPO=malice
NAME=engine
VERSION=$(shell cat .release/VERSION 2>/dev/null || echo "0.3.28")
MESSAGE?="New release"

# Modern Go project settings
GO := go
GOFLAGS := -v
GOTEST := go test

GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY := $(shell git status --porcelain 2>/dev/null | grep -q . && echo "+CHANGES" || echo "")
GIT_DESCRIBE := $(shell git describe --tags 2>/dev/null || echo "v0.3.28")

# Targets
.DEFAULT_GOAL := help

setup: ## Setup development environment - install dependencies
	@echo "===> Setting up development environment for Go 1.21+"
	$(GO) version
	$(GO) mod tidy
	$(GO) mod download

fmt: ## Format Go code
	@echo "===> Formatting Go code"
	$(GO) fmt ./...
	goimports -w .

lint: ## Run linters
	@echo "===> Running linters"
	golangci-lint run ./... || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

build: ## Build the malice binary
	@echo "===> Building Malice binary"
	$(GO) build -v \
		-ldflags "-s -w \
			-X main.version=$(VERSION) \
			-X main.commit=$(GIT_COMMIT) \
			-X main.date=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')" \
		-o build/malice ./

test: ## Run tests
	@echo "===> Running tests"
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test ## Run tests and generate coverage report
	@echo "===> Generating coverage report"
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "===> Cleaning build artifacts"
	$(GO) clean
	rm -rf build/ dist/ coverage.txt coverage.html

docker-build: ## Build Docker image for Ubuntu 22.04
	@echo "===> Building Docker image"
	docker build -t $(REPO)/$(NAME):$(VERSION) -f Dockerfile .

docker-run: docker-build ## Run Docker container
	@echo "===> Running Docker container"
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock $(REPO)/$(NAME):$(VERSION)

run: build ## Run the malice binary
	@echo "===> Running Malice"
	./build/malice -D

install: build ## Install the binary to GOPATH/bin
	@echo "===> Installing Malice"
	$(GO) install

release: clean build ## Create a release
	@echo "===> Creating release $(VERSION)"
	mkdir -p dist
	cp build/malice dist/malice-$(VERSION)-linux-amd64
	@echo "Release ready: dist/malice-$(VERSION)-linux-amd64"

deps: ## Show dependencies
	@$(GO) list -m all

deps-update: ## Update all dependencies to latest compatible versions
	@echo "===> Updating dependencies"
	$(GO) get -u ./...
	$(GO) mod tidy

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.DEFAULT_GOAL := help
