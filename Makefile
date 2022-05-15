GO=GO111MODULE=on go
WIRE=wire
COVERAGE_FILE="/tmp/copper_coverage.out"

.PHONY: all
all: lint test

.PHONY: cover
cover: test
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: test
test:
	$(GO) test -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: lint
lint: tidy
	golangci-lint run

.PHONY: tidy
tidy:
	$(GO) mod tidy

