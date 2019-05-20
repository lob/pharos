BIN_DIR         ?= ./bin
PKG_SERVER_NAME ?= pharos-api-server
PKG_CLI_NAME    ?= pharos

COVERAGE_PROFILE ?= coverage.out

GOTOOLS := \
	github.com/codegangsta/gin \
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \

default: build

.PHONY: build
build:
	@echo "---> Building"
	go build -o $(BIN_DIR)/$(PKG_SERVER_NAME) ./cmd/pharos-api-server
	go build -o $(BIN_DIR)/$(PKG_CLI_NAME) ./cmd/pharos

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

.PHONY: setup
setup:
	@echo "--> Installing development tools"
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_DIR) v1.15.0
	go get -u $(GOTOOLS)

.PHONY: start
start:
	@echo "---> Starting"
	gin --path . --build ./cmd/pharos-api-server --immediate --bin $(BIN_DIR)/gin-$(PKG_NAME) run

.PHONY: test
test:
	@echo "---> Testing"
	ENVIRONMENT=test go test ./pkg/... -race -coverprofile $(COVERAGE_PROFILE)
