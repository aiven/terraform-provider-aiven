GO := CGO_ENABLED=0 go

#################################################
# Tools
#################################################

TOOLS_DIR := tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin


$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)


TFPLUGINDOCS := $(TOOLS_BIN_DIR)/tfplugindocs

$(TFPLUGINDOCS): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/tfplugindocs github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs


TERRAFMT := $(TOOLS_BIN_DIR)/terrafmt

$(TERRAFMT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/terrafmt github.com/katbyte/terrafmt


GOLANGCILINT := $(TOOLS_BIN_DIR)/golangci-lint

$(GOLANGCILINT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

#################################################
# Build
#################################################

# See https://github.com/hashicorp/terraform/blob/main/tools/protobuf-compile/protobuf-compile.go#L215
ARCH := $(shell $(GO) env GOOS GOARCH | tr '\n' '_' | sed '$$s/_$$//')
BUILD_DEV_DIR := ~/.terraform.d/plugins/registry.terraform.io/aiven/aiven/0.0.0+dev/$(ARCH)


$(BUILD_DEV_DIR):
	mkdir -p $(BUILD_DEV_DIR)


build:
	$(GO) build


# Example usage in Terraform configuration:
#
# terraform {
#  required_providers {
#    aiven = {
#      source  = "aiven/aiven"
#      version = "0.0.0+dev"
#    }
#  }
#}
build-dev: $(BUILD_DEV_DIR)
	$(GO) build -o $(BUILD_DEV_DIR)/terraform-provider-aiven_v0.0.0+dev

#################################################
# Test
#################################################

.PHONY: test
test: test-unit test-acc


.PHONY: test-unit
test-unit:
	$(GO) test -v --cover ./...


PKG ?= internal
ifneq ($(origin PKG), file)
	PKG := internal/service/$(PKG)
endif

TEST_COUNT := 1
ACC_TEST_TIMEOUT := 180m
ACC_TEST_PARALLELISM := 20

.PHONY: test-acc
test-acc:
	TF_ACC=1 $(GO) test ./$(PKG)/... \
	-v -count $(TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACC_TEST_TIMEOUT)

#################################################
# Lint
#################################################

.PHONY: lint
lint: lint-go lint-test lint-docs


.PHONY: lint-go
lint-go: $(GOLANGCILINT)
	$(GOLANGCILINT) run --timeout=30m ./...


.PHONY: lint-test
lint-test: $(TERRAFMT)
	$(TERRAFMT) diff ./internal -cfv


.PHONY: lint-docs
lint-docs: $(TFPLUGINDOCS)
	$(TFPLUGINDOCS) validate

#################################################
# Format
#################################################

.PHONY: fmt
fmt: fmt-test

.PHONY: fmt-test
fmt-test: $(TERRAFMT)
	$(TERRAFMT) fmt ./internal -fv

#################################################
# Docs
#################################################

.PHONY: docs
docs: $(TFPLUGINDOCS)
	$(TFPLUGINDOCS) generate

#################################################
# Misc
#################################################

.PHONY: clean
clean: clean-tools sweep


.PHONY: clean-tools
clean-tools:
	rm -rf $(TOOLS_BIN_DIR)


SWEEP := global
SWEEP_DIR := ./internal/sweep

.PHONY: sweep
sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	$(GO) test $(SWEEP_DIR) -v -tags=sweep -sweep=$(SWEEP) $(SWEEP_ARGS) -timeout 15m
