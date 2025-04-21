NAME = nurd-commerce

SWAGGER_YAML = docs/swagger/swagger.yml
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

MIGRATION_FILE = $(shell date +"migrations/%Y%m%d%H%M%S-$(name).sql")
DATA_MIGRATION_FILE = $(shell date +"data-migrations/%Y%m%d%H%M%S-$(name).sql")

.PHONY: help \
	lint lint-fix \
	fmt env run-dev run-worker \
	setup setup-covmerge setup-goimports setup-migrate setup-mockgen setup-security setup-docs setup-lint \
	test coverage get-coverage mocks \
	migrate new-migration new-data-migration \
	start-env start-app start-monitoring start-all \
	stop-env stop-app \
	build-docker build-container \
	connect-db create-volume security-check

clean:
	rm -rf $(NAME)
	rm -rf $(COVERAGE_FILE)

## Install all the build and lint dependencies
setup: setup-covmerge setup-goimports setup-migrate setup-mockgen setup-security setup-docs setup-lint setup-sqlfluff

setup-lint: ## Install the linter
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

setup-goimports: ## Install the goimports
	go install -mod=mod golang.org/x/tools/cmd/goimports@latest

setup-covmerge: ## Install the covmerge tool
	go install -mod=mod github.com/wadey/gocovmerge

setup-migrate: ## Install the migrate tool
	go install -mod=mod github.com/nurdsoft/nurd-commerce-core/shared/db/migrate

setup-mockgen: ## Install mockgen to generate mocks
	go install github.com/golang/mock/mockgen@latest

setup-security: ## Install govulncheck to check for vulnerabilities
	go install golang.org/x/vuln/cmd/govulncheck@latest

setup-docs:
	go install github.com/go-swagger/go-swagger/cmd/swagger@latest

setup-sqlfluff:
	docker pull sqlfluff/sqlfluff:latest

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

lint:
	golangci-lint run

lint-sql:
	docker run -it --rm -v $$PWD:/sql sqlfluff/sqlfluff:latest lint migrations/ --dialect postgres

lint-fix:
	golangci-lint run --fix

$(NAME): docs
	go generate ./...
	go run scripts/error_extractor.go
	go build -race -o $(NAME) .

env:
	@echo "Exporting environment variables"
	$(shell export $$(grep --color=never -v '^#' .env | xargs))
	@echo "Done"

run-dev: $(NAME)
	@export $(shell grep --color=never -v '^#' .env | xargs) && ./$(NAME) api

test:
	go test -race ./... -coverpkg=./... -coverprofile=$(COVERAGE_FILE)

coverage:
	go run scripts/coverage_filter.go $(COVERAGE_FILE)
	go tool cover -html $(COVERAGE_FILE) -o $(COVERAGE_HTML)

get-coverage: ## Get overall project test coverage
	go run scripts/get_overall_coverage.go

mocks: ## run custom script to update mocks
	go run scripts/mock_updater.go

migrate: env ## Apply outstanding migrations
	./$(NAME) migrate --direction=$(direction)

new-migration: ## New migration (make name=add-some-table new-migration)
	touch $(MIGRATION_FILE)
	echo "-- +migrate Up\n\n-- +migrate Down" >> $(MIGRATION_FILE)

new-data-migration: ## New migration (make name=add-some-table new-migration)
	touch $(DATA_MIGRATION_FILE)
	echo "-- +migrate Up\n\n-- +migrate Down" >> $(DATA_MIGRATION_FILE)

start-env: ## Start the local env
	docker-compose up -d otel-collector
	docker-compose up -d db

start-app: ## Start the application in docker container
	docker-compose up --build api

stop-app: ## Stop db and api
	docker-compose stop api db

start-monitoring: ## Start the monitoring services in docker container
	docker-compose up -d --build prometheus grafana

start-all: ## Start the environment services and the application in docker container
	docker-compose up -d

stop-env: ## Stop the local env
	docker-compose down

build-docker: ## Build docker env
	docker-compose build

connect-db: ## Connect to redesign local db
	psql postgresql://db:123@localhost:5452/commerce-core

create-volume:
	docker volume create --driver local --opt type=tmpfs --opt device=tmpfs --opt o=size=$(STORAGE)m,uid=1000 api-storage

build-container:
	docker build -t $(NAME):local .

run-container:
	docker run -v api-storage:/go --network=nurdsoft-comerce_core -p 8080:8080 -e AWS_ACCESS_KEY_ID="$$AWS_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$$AWS_SECRET_ACCESS_KEY" -e AWS_SESSION_TOKEN="$$AWS_SESSION_TOKEN" $(NAME):local

security-check:
	govulncheck ./...

docs: docs/swagger/swagger.yml

docs/swagger/swagger.yml:
	swagger generate spec -o ./docs/swagger/swagger.yml --scan-models

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
