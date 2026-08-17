BINARY   := bin/api
PKG      := ./...
COVERAGE := coverage.out

DB_URL := postgres://scoreplay:scoreplay@postgres:5432/scoreplay?sslmode=disable
MIGRATE := docker compose run --rm migrate -path=/migrations -database='$(DB_URL)'

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

# --- compose stack ------------------------------------------------------------------

.PHONY: psql
psql: ## Open a psql shell in the running database
	docker compose exec postgres psql -U scoreplay -d scoreplay

# --- migrations ---------------------------------------------------------------------

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(MIGRATE) up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration (local resets only)
	$(MIGRATE) down 1

.PHONY: migrate-create
migrate-create: ## Scaffold a migration pair: make migrate-create NAME=add_something
	@test -n "$(NAME)" || (echo "NAME is required, e.g. make migrate-create NAME=add_processing_status" && exit 1)
	@last=$$(ls migrations/*.up.sql 2>/dev/null | sed -E 's#.*/0*([0-9]+)_.*#\1#' | sort -n | tail -1); \
	 next=$$(printf '%06d' $$(( $${last:-0} + 1 ))); \
	 touch "migrations/$${next}_$(NAME).up.sql" "migrations/$${next}_$(NAME).down.sql"; \
	 echo "created migrations/$${next}_$(NAME).{up,down}.sql"

.PHONY: migrate-force
migrate-force: ## Clear a dirty migration state: make migrate-force VERSION=1
	@test -n "$(VERSION)" || (echo "VERSION is required, e.g. make migrate-force VERSION=1" && exit 1)
	$(MIGRATE) force $(VERSION)
