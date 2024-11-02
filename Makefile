# A phony target is one that is not really the name of a file; 
# rather it is just a name for a rule to be executed.

include .env 

# ============================================================================= #
# HELPERS
# ============================================================================= #

## help: print this help msg
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

# ============================================================================= #
# DEVELOPMENT
# ============================================================================= #

## run/api: connect to the cmd/api application
.PHONY: run/api
run/api:
	# @go run ./cmd/api # use @ to supress command echo
	go run ./cmd/api 

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${db_dsn}

## db/migrations/new name=$1: create a new database migration
#run this using make db/migrations/new name=<arg>
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up databse migrations
# depends on confirm
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migration...'
	migrate -path ./migrations -database ${db_dsn} up

# ============================================================================= #
# QUALITY CONTROl
# ============================================================================= #

## audit: tidy dependencies and format, vet and test
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ============================================================================= #
# BUILD
# ============================================================================= #

linux_current_time = $(shell date --iso-8601=seconds)
macOS_current_time = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

git_description = ${shell git describe --always --dirty --tags --long}

# ## build/api/macos: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s -X main.buildTime=${macOS_current_time} -X main.version=${git_description}' -o=./bin/api ./cmd/api
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags='-s -X main.buildTime=${linux_current_time} -X main.version=${git_description}' -o=./bin/${GOOS}_${GOARCH}/api ./cmd/api


# ## build/api/macos: build the cmd/api application for macOS
# .PHONY: build/api/macos
# build/api/macos:
# 	@echo 'Building cmd/api for macOS...'
# 	go build -ldflags='-s -X main.buildTime=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")' -o=./bin/api ./cmd/api

# ## build/api: build the cmd/api application
# .PHONY: build/api/linux
# build/api/linux:
# 	@echo 'Building cmd/api for Linux...'
# 	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags='-s -X main.buildTime=$(shell date --iso-8601=seconds)' -o=./bin/${GOOS}_${GOARCH}/api ./cmd/api