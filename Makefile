version = $(shell git describe --long --tags 2>/dev/null || echo unknown-g`git describe --always`)
short_version = $(shell echo $(version) | sed 's/-.*//')

GO=CGO_ENABLED=0 go
BUILDFLAGS=-ldflags "-X main.version=${version}" 

#################################################
# Tools
#################################################

TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

TFPLUGINDOCS=$(TOOLS_BIN_DIR)/tfplugindocs

$(TFPLUGINDOCS): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod ## Build tfplugindocs from tools folder.
	cd $(TOOLS_DIR) && $(GO) build -o bin/tfplugindocs github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

#################################################
# Docs
#################################################

.PHONY: docs
docs: $(TFPLUGINDOCS)
	$(TFPLUGINDOCS) generate

.PHONY: docs-lint
docs-lint: $(TFPLUGINDOCS)
	$(TFPLUGINDOCS) validate


#################################################
# Testing and linting
#################################################

.PHONY: test
test:
	CGO_ENABLED=0 go test -v --cover ./...

.PHONY: testacc
testacc:
	TF_ACC=1 CGO_ENABLED=0 go test -v -count 1 -parallel 30 --cover ./... -timeout 120m ${TESTARGS}

.PHONY: sweep
sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test -v ./aiven -sweep=global -timeout 60m

.PHONY: lint
lint:
	golangci-lint run --issues-exit-code=0 --timeout=30m ./...
