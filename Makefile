BINARY   := bin/api
PKG      := ./...
COVERAGE := coverage.out

.DEFAULT_GOAL := help

.PHONY: help
help: ## List available targets
	@awk 'BEGIN { FS = ":.*##"; printf "\nUsage: make <target>\n\n" } \
		/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } \
		END { printf "\n" }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Compile the API binary into bin/
	go build -o $(BINARY) ./cmd/api

.PHONY: run
run: ## Run the API locally (reads .env if exported)
	go run ./cmd/api

.PHONY: test
test: ## Run unit tests with the race detector
	go test -race $(PKG)

.PHONY: test-integration
test-integration: ## Run integration tests (requires Docker)
	go test -race -tags=integration -count=1 $(PKG)

.PHONY: cover
cover: ## Run unit tests and open the coverage report
	go test -coverprofile=$(COVERAGE) $(PKG)
	go tool cover -html=$(COVERAGE)

.PHONY: fmt
fmt: ## Format all Go code
	gofmt -w .

.PHONY: lint
lint: ## Run go vet and golangci-lint
	go vet $(PKG)
	golangci-lint run

.PHONY: tidy
tidy: ## Sync go.mod/go.sum with the source
	go mod tidy

.PHONY: clean
clean: ## Remove build and coverage artifacts
	rm -rf bin $(COVERAGE)
