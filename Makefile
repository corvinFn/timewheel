GO ?= go
GOFMT ?= gofmt "-s"
GO_ENV ?= local
VERSION ?= $(shell git describe --tags --always || git rev-parse --short HEAD)
IMAGE_TAG ?= $(GO_ENV)-$(VERSION)
APP_TAG ?= .APP_TAG

GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: lint lint_install lint_check
lint: lint_install lint_check
	@$(GOBIN)/golangci-lint version
	@$(GOBIN)/golangci-lint run

LINTVERSION ?= v1.50.1
GOBIN = $(HOME)/go/bin
LINTINSTALL = $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINTVERSION)
lint_install:
	@if [ ! -d $(GOBIN) ] ; then mkdir -p $(GOBIN) ; else echo "GOBIN: $(GOBIN)" ; fi
	@if [ ! -f $(GOBIN)/golangci-lint ] ; then GOBIN=$(GOBIN) $(LINTINSTALL); else echo "golangci-lint exists" ; fi

CURRENT_VERSION := `$(GOBIN)/golangci-lint version | sed -nre 's/^[^0-9]*(([0-9]+\.)*[0-9]+).*/\1/p'`
lint_check:
	@if [ "$(shell echo v$(CURRENT_VERSION))" != "$(LINTVERSION)" ] ; then GOBIN=$(GOBIN) $(LINTINSTALL) ; fi

.PHONY: test
test:
	$(GO) test -race -v -p 4 -race -cover -coverprofile=cover.out ./...
