BIN_DIR         ?= ./bin
PKG_SERVER_NAME ?= pharos-api-server
PKG_CLI_NAME    ?= pharos
LDFLAGS         ?= "-X github.com/lob/pharos/pkg/pharos/cmd.pharosVersion=$(VERSION)"
VERSION         ?=

COVERAGE_PROFILE ?= coverage.out

GOTOOLS := \
	github.com/codegangsta/gin \
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \

PSQL := $(shell command -v psql 2> /dev/null)

DATABASE_USER             ?= pharos_admin
DATABASE_NAME_DEVELOPMENT ?= pharos
DATABASE_NAME_TEST        ?= pharos_test

default: build

.PHONY: build
build:
	@echo "---> Building"
	go build -o $(BIN_DIR)/$(PKG_SERVER_NAME) ./cmd/pharos-api-server
	go build -o $(BIN_DIR)/$(PKG_CLI_NAME) -ldflags $(LDFLAGS) ./cmd/pharos

.PHONY: clean
clean:
	@echo "---> Cleaning"
	rm -rf $(BIN_DIR) ./vendor

.PHONY: enforce
enforce:
	@echo "---> Enforcing coverage"
	./scripts/coverage.sh $(COVERAGE_PROFILE)

.PHONY: html
html:
	@echo "---> Generating HTML coverage report"
	go tool cover -html $(COVERAGE_PROFILE)

.PHONY: install
install:
	@echo "---> Installing dependencies"
	dep ensure -vendor-only

.PHONY: lint
lint:
	@echo "---> Linting"
	$(BIN_DIR)/golangci-lint run

.PHONY: migrate
migrate:
	@echo "---> Migrating"
	go run cmd/migrations/*.go migrate

.PHONY: rollback
rollback:
	@echo "---> Rolling back"
	go run cmd/migrations/*.go rollback

.PHONY: setup
setup:
	@echo "--> Installing development tools"
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_DIR) v1.16.0
	go get -u $(GOTOOLS)
ifdef PSQL
	dropdb --if-exists $(DATABASE_NAME_DEVELOPMENT)
	dropdb --if-exists $(DATABASE_NAME_TEST)
	dropuser --if-exists $(DATABASE_USER)
	createuser --createdb $(DATABASE_USER)
	createdb -U $(DATABASE_USER) $(DATABASE_NAME_DEVELOPMENT)
	psql -U $(DATABASE_USER) -d $(DATABASE_NAME_DEVELOPMENT) -c "ALTER DATABASE $(DATABASE_NAME_DEVELOPMENT) SET timezone TO 'UTC'"
	createdb -U $(DATABASE_USER) $(DATABASE_NAME_TEST)
	psql -U $(DATABASE_USER) -d $(DATABASE_NAME_TEST) -c "ALTER DATABASE $(DATABASE_NAME_TEST) SET timezone TO 'UTC'"
	make install
	make migrate
	ENVIRONMENT=test make migrate
else
	@echo "Skipping database setup"
endif

.PHONY: start
start:
	@echo "---> Starting"
	gin --path . --build ./cmd/pharos-api-server --immediate --bin $(BIN_DIR)/gin-$(PKG_NAME) run

.PHONY: test
test:
	@echo "---> Testing"
	ENVIRONMENT=test go test ./pkg/... -race -coverprofile $(COVERAGE_PROFILE)

.PHONY: release
release:
	@echo "---> Creating tagged release"
ifndef VERSION
	$(error VERSION must be specified)
endif
	git tag $(VERSION)
	# Check that the commit is tagged and starts with "v".
	[[ $$(git tag -l --points-at HEAD) == v* ]]
	git push origin
	git push origin --tags
