MIGRATION_FILE = $(shell date +"migrations/%Y%m%d%H%M%S-$(name).sql")
DATA_MIGRATION_FILE = $(shell date +"data-migrations/%Y%m%d%H%M%S-$(name).sql")
RAM = 64
STORAGE = 300
AWS_REGISTRY ?= nurdsoft
AWS_REPOSITORY ?= api
IMAGE_TAG ?= v0.1.0

setup-lint: ## Install the linter
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

setup-goimports: ## Install the goimports
	go install -mod=mod golang.org/x/tools/cmd/goimports@latest

setup-covmerge: ## Install the covmerge tool
	go get github.com/wadey/gocovmerge
	go install -mod=mod github.com/wadey/gocovmerge

setup-migrate: ## Install the migrate tool
	go install -mod=mod github.com/nurdsoft/nurd-commerce-core/shared/db/migrate

setup-mockgen: ## Install mockgen to generate mocks
	go install github.com/golang/mock/mockgen@latest

setup-security: ## Install govulncheck to check for vulnerabilities
	go install golang.org/x/vuln/cmd/govulncheck@latest

setup-docs:
	go install github.com/go-swagger/go-swagger/cmd/swagger@latest

setup: setup-covmerge setup-goimports setup-migrate setup-mockgen setup-security setup-lint setup-docs ## Install all the build and lint dependencies

dep: ## Get all dependencies
	go env -w GOPROXY=direct
	go env -w GOSUMDB=off
	go mod download
	go mod tidy

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

lint: dep ## Run linter for the code
	golangci-lint run

cleanup-lint: dep ## Run all the linters and clean up
	golangci-lint run --fix

auth-ecr: ## AWS ECR Authentication 
	aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin ${AWS_REGISTRY}

build-dev: dep ## Build a beta version
	go generate ./...
	go run scripts/error_extractor.go
	go build -race -o commerce-core .

build-docker: ## Build docker env
	docker-compose build

env:
	@echo "Exporting environment variables"
	$(shell export $$(grep --color=never -v '^#' .env | xargs))
	@echo "Done"

run-dev: build-dev ## Run the application locally
	@export $(shell grep --color=never -v '^#' .env | xargs) && ./commerce-core api

run-worker: build-dev ## Run the worker locally
	./commerce-core worker

test: ## Run tests
	go test -race ./... -coverpkg=./... -coverprofile=coverage.out

coverage:
	go run scripts/coverage_filter.go coverage.out
	go tool cover -html coverage.out -o coverage.html

get-coverage: ## Get overall project test coverage
	go run scripts/get_overall_coverage.go

mocks: ## run custom script to update mocks
	go run scripts/mock_updater.go

migrate: env ## Apply outstanding migrations
	./commerce-core migrate --direction=$(direction)

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

build-image: ## Build Docker Image
	docker build -t $(AWS_REGISTRY)/$(AWS_REPOSITORY):$(IMAGE_TAG) .

push-image: ## Push Image to Registry
	docker push $(AWS_REGISTRY)/$(AWS_REPOSITORY):$(IMAGE_TAG)

cleanup-build: ## Cleanup executable
	rm -f commerce-core

cleanup: cleanup-build ## Cleanup all files

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

connect-db: ## Connect to redesign local db
	psql postgresql://db:123@localhost:5452/commerce-core

create-volume:
	docker volume create --driver local --opt type=tmpfs --opt device=tmpfs --opt o=size=$(STORAGE)m,uid=1000 api-storage

build-container:
	docker build -t commerce-core:local .

run-container:
	docker run -v api-storage:/go --network=nurdsoft-comerce_core -p 8080:8080 -e AWS_ACCESS_KEY_ID="$$AWS_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$$AWS_SECRET_ACCESS_KEY" -e AWS_SESSION_TOKEN="$$AWS_SESSION_TOKEN" commerce-core:local

bnd-container:
	docker run -v api-storage:/go --memory $(RAM)m --network=nurdsoft-comerce_core -p 8080:8080 -e AWS_ACCESS_KEY_ID="$$AWS_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$$AWS_SECRET_ACCESS_KEY" -e AWS_SESSION_TOKEN="$$AWS_SESSION_TOKEN" commerce-core:local

security-check:
	govulncheck ./...

generate-docs:
	swagger generate spec -o ./docs/swagger/swagger.yml --scan-models

.DEFAULT_GOAL := help
