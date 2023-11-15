.PHONY: get lint build install

GO ?= go
GOLINT ?= golangci-lint

GO_BUILD_ARGS ?= -v
GO_INSTALL_ARGS ?= -v
GO_GENERATE_ARGS ?= -v
GO_GET_ARGS ?= -v

GOLINT_URL ?= github.com/golangci/golangci-lint/cmd/...
GOLINT_VERSION ?= latest
GOLINT_ARGS ?=

build:
	$(GO) generate $(GO_GENERATE_ARGS) ./...
	$(GO) build $(GO_BUILD_ARGS) ./...

install:
	$(GO) install $(GO_BUILD_ARGS) $(GO_INSTALL_ARGS) ./cmd/...

get:
	$(GO) get $(GO_GET_ARGS) ./...
	which $(GOLINT) > /dev/null || \
		$(GO) install $(GO_INSTALL_ARGS) $(GOLINT_URL)@$(GOLINT_VERSION)

lint:
	which $(GOLINT) > /dev/null || \
		$(GO) install $(GO_INSTALL_ARGS) $(GOLINT_URL)@$(GOLINT_VERSION)
	$(GOLINT) $(GOLINT_ARGS) run --fix
