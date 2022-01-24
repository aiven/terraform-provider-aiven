version = $(shell git describe --long --tags 2>/dev/null || echo unknown-g`git describe --always`)
short_version = $(shell echo $(version) | sed 's/-.*//')

GO=CGO_ENABLED=0 go
BUILDFLAGS=-ldflags "-X main.version=${version}" 

SOURCES = $(shell find aiven -name '*.go')

#################################################
# Tools
#################################################

TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

TFPLUGINDOCS=$(TOOLS_BIN_DIR)/tfplugindocs

$(TFPLUGINDOCS): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/tfplugindocs github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

TFPROVIDERTESTFMT=$(TOOLS_BIN_DIR)/tfprovidertestfmt

$(TFPROVIDERTESTFMT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/tfprovidertestfmt github.com/aiven/tfprovidertestfmt

GOLANGCILINT=$(TOOLS_BIN_DIR)/golangci-lint

$(GOLANGCILINT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

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
lint: go-lint test-lint docs-lint

.PHONY: go-lint
go-lint: $(GOLANGCILINT)
	$(GOLANGCILINT) run --timeout=30m ./...

.PHONY: test-lint
test-lint: $(TFPROVIDERTESTFMT)
	$(TFPROVIDERTESTFMT) -lint $(SOURCES)

.PHONY: testfmt
testfmt: $(TFPROVIDERTESTFMT)
	$(TFPROVIDERTESTFMT) -inplace $(SOURCES)


#################################################
# Misc
#################################################

clean: clean-tools

clean-tools:
	rm -rf $(TOOLS_BIN_DIR)
