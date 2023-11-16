.PHONY: all get generate lint build install

# go
#
GO ?= go
GO_BUILD_ARGS ?= -v
GO_INSTALL_ARGS ?=
GO_GENERATE_ARGS ?= -v
GO_GET_ARGS ?= -v

GO_BUILD := $(GO) build $(GO_BUILD_ARGS)
GO_INSTALL := $(GO) install $(GO_BUILD_ARGS) $(GO_INSTALL_ARGS)
GO_GENERATE := $(GO) generate $(GO_GENERATE_ARGS)
GO_GET := $(GO) get $(GO_GET_ARGS)

# golint
#
GOLINT ?= golangci-lint
GOLINT_URL ?= github.com/golangci/golangci-lint/cmd/...
GOLINT_VERSION ?= latest
GOLINT_ARGS ?=
GOLINT_RUN ?= run
GOLINT_RUN_ARGS ?= --fix


GO_INSTALL_DEPS =
ifeq ($(shell which $(GOLINT)),)
GO_INSTALL_DEPS += $(GOLINT_URL)@$(GOLINT_VERSION)
endif

all: get lint build

deps:
	@for u in $(GO_INSTALL_DEPS); do \
		$(GO_INSTALL) $$u; \
	done

generate: deps
	@find * -type f | xargs -r grep -l '^//go:generate' \
		| xargs -rtl $(GO_GENERATE)

build: generate
	$(GO_BUILD) ./...

install: deps
	$(GO_INSTALL) ./cmd/...

get: deps
	$(GO_GET) ./...

lint: deps
	$(GO) mod tidy
	$(GOLINT) $(GOLINT_ARGS) $(GOLINT_RUN) $(GOLINT_RUN_ARGS)
