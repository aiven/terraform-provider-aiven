version = $(shell git describe --long --tags 2>/dev/null || echo unknown-g`git describe --always`)
short_version = $(shell echo $(version) | sed 's/-.*//')

GO=CGO_ENABLED=0 go
BUILDFLAGS=-ldflags "-X main.version=${version}" 

.PHONY: ci
ci: lint bins release

#################################################
# Building
#################################################

.PHONY: plugins
plugins:
	mkdir -p plugins/linux_amd64 plugins/darwin_amd64

.PHONY: bins
bins: plugins
	$(GO) generate
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o plugins/linux_amd64/terraform-provider-aiven_$(short_version) .
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o plugins/darwin_amd64/terraform-provider-aiven_$(short_version) .
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o plugins/windows_amd64/terraform-provider-aiven_$(short_version).exe .
	GOOS=windows GOARCH=386 $(GO) build $(BUILDFLAGS) -o plugins/windows_386/terraform-provider-aiven_$(short_version).exe .

#################################################
# Tools
#################################################

TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

TFPLUGINDOCS=$(TOOLS_BIN_DIR)/tfplugindocs
ACT=$(TOOLS_BIN_DIR)/act

$(TFPLUGINDOCS): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod ## Build tfplugindocs from tools folder.
	cd $(TOOLS_DIR) && $(GO) build -o bin/tfplugindocs github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

$(ACT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod ## Build act from tools folder.
	cd $(TOOLS_DIR) && $(GO) build -o bin/act github.com/nektos/act

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

#################################################
# Artifacts for release
#################################################

.PHONY: release
release: bins
	tar cvzf terraform-provider-aiven.tar.gz -C plugins \
	    linux_amd64/terraform-provider-aiven_$(short_version) \
	    darwin_amd64/terraform-provider-aiven_$(short_version) \
	    windows_amd64/terraform-provider-aiven_$(short_version).exe \
	    windows_386/terraform-provider-aiven_$(short_version).exe


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

.PHONY: testacc
testacc:
	TF_ACC=1 CGO_ENABLED=0 go test -v -count 1 -parallel 10 --cover ./... $(TESTARGS) -timeout 120m

.PHONY: sweep
sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test -v ./aiven -sweep=global -timeout 60m

.PHONY: lint
lint:
	golangci-lint run --issues-exit-code=0 --timeout=30m ./...

.PHONY: clean
clean:
	rm -rf vendor
	rm -rf plugins
	rm -f terraform-provider-aiven.tar.gz

#################################################
# Local CI Workflows
#################################################

ACTARGS=--secret-file=aiven.secret
.PHONY: github-actions-acceptance-tests
github-actions-acceptance-tests: $(ACT)
	$(ACT) $(ACTARGS) -j acceptance_tests

.PHONY: github-actions-unit-tests
github-actions-unit-tests: $(ACT)
	$(ACT) $(ACTARGS) -j go_test

.PHONY: github-actions-golangci-lint
github-actions-golangci-lint: $(ACT)
	$(ACT) $(ACTARGS) -j golangci_lint

