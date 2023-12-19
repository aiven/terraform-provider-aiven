.PHONY: build build-dev debug test test-unit test-acc test-examples lint lint-go lint-test lint-docs fmt fmt-test fmt-imports clean clean-tools clean-examples sweep generate gen-go docs ci-selproj

#################################################
# Global
#################################################

GO := CGO_ENABLED=0 go

TOOLS_DIR ?= tools
TOOLS_BIN_DIR ?= $(TOOLS_DIR)/bin

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)


GOLANGCILINT := $(TOOLS_BIN_DIR)/golangci-lint

$(GOLANGCILINT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint


TFPLUGINDOCS := $(TOOLS_BIN_DIR)/tfplugindocs

$(TFPLUGINDOCS): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/tfplugindocs github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs


TERRAFMT := $(TOOLS_BIN_DIR)/terrafmt

$(TERRAFMT): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o bin/terrafmt github.com/katbyte/terrafmt


SELPROJ := $(TOOLS_BIN_DIR)/selproj

$(SELPROJ): $(TOOLS_BIN_DIR) $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -tags tools -o bin/selproj github.com/aiven/terraform-provider-aiven/tools/selproj


# See https://github.com/hashicorp/terraform/blob/main/tools/protobuf-compile/protobuf-compile.go#L215
ARCH ?= $(shell $(GO) env GOOS GOARCH | tr '\n' '_' | sed '$$s/_$$//')
BUILD_DEV_DIR ?= ~/.terraform.d/plugins/registry.terraform.io/aiven-dev/aiven/0.0.0+dev/$(ARCH)
BUILD_DEV_BIN ?= $(BUILD_DEV_DIR)/terraform-provider-aiven_v0.0.0+dev

$(BUILD_DEV_DIR):
	mkdir -p $(BUILD_DEV_DIR)

#################################################
# Build
#################################################

build:
	$(GO) build


# Example usage in Terraform configuration:
#
# terraform {
#  required_providers {
#    aiven = {
#      source  = "aiven-dev/aiven"
#      version = "0.0.0+dev"
#    }
#  }
#}
build-dev: $(BUILD_DEV_DIR)
	$(GO) build -gcflags='all=-N -l' -o $(BUILD_DEV_BIN)

#################################################
# Debug
#################################################

debug: build-dev
	$(BUILD_DEV_BIN) -debug

#################################################
# Test
#################################################

test: test-unit test-acc


test-unit:
	$(GO) test -v --cover ./...


PKG_PATH ?= internal
ifneq ($(origin PKG), undefined)
	PKG_PATH = internal/sdkprovider/service/$(PKG)
endif

TEST_COUNT ?= 1
ACC_TEST_TIMEOUT ?= 180m
ACC_TEST_PARALLELISM ?= 10

test-acc:
	TF_ACC=1 $(GO) test ./$(PKG_PATH)/... \
	-v -count $(TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACC_TEST_TIMEOUT)


test-examples: build-dev clean-examples
	AIVEN_PROVIDER_PATH=$(BUILD_DEV_DIR) $(GO) test --tags=examples ./examples_tests/... \
	-v -count $(TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACC_TEST_TIMEOUT)

#################################################
# Lint
#################################################

lint: lint-go lint-test lint-docs


lint-go: $(GOLANGCILINT)
	$(GOLANGCILINT) run --build-tags all --timeout=30m ./...


lint-test: $(TERRAFMT)
	$(TERRAFMT) diff ./internal -cfq


lint-docs: $(TFPLUGINDOCS)
	PROVIDER_AIVEN_ENABLE_BETA=true $(TFPLUGINDOCS) validate

#################################################
# Format
#################################################

fmt: fmt-test fmt-imports


fmt-test: $(TERRAFMT)
	$(TERRAFMT) fmt ./internal -fv


# macOS requires to install GNU sed first. Use `brew install gnu-sed` to install it.
# It has to be added to PATH as `sed` command, to replace default BSD sed.
# See `brew info gnu-sed` for more details on how to add it to PATH.
fmt-imports:
	find . -type f -name '*.go' -exec sed -zi 's/(?<== `\s+)"\n\+\t"/"\n"/g' {} +
	goimports -local "github.com/aiven/terraform-provider-aiven" -w .

#################################################
# Clean
#################################################

clean: clean-tools clean-examples sweep


clean-tools: $(TOOLS_BIN_DIR)
	rm -rf $(TOOLS_BIN_DIR)


clean-examples:
	find ./examples -type f -name '*.tfstate*' -delete


SWEEP ?= global

sweep:
	@echo 'WARNING: This will destroy infrastructure. Use only in development accounts.'
	$(GO) test ./internal/sweep -v -tags=sweep -sweep=$(SWEEP) $(SWEEP_ARGS) -timeout 15m

#################################################
# Generate
#################################################

generate: gen-go docs


gen-go:
	go generate ./...
	$(MAKE) fmt-imports


docs: $(TFPLUGINDOCS)
	PROVIDER_AIVEN_ENABLE_BETA=true $(TFPLUGINDOCS) generate

#################################################
# CI
#################################################

ci-selproj: $(SELPROJ)
	$(SELPROJ)
