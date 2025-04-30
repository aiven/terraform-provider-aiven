.PHONY: build build-dev debug test test-unit test-acc test-examples lint lint-go lint-test lint-docs fmt fmt-test fmt-imports clean clean-tools clean-examples sweep generate gen-go docs ci-selproj

#################################################
# Global
#################################################

GO := CGO_ENABLED=0 go
GOTOOL := $(GO) tool
CONTAINER_TOOL ?= podman

# See https://github.com/hashicorp/terraform/blob/main/tools/protobuf-compile/protobuf-compile.go#L215
ARCH ?= $(shell $(GO) env GOOS GOARCH | tr '\n' '_' | sed '$$s/_$$//')
BUILD_DEV_DIR ?= ~/.terraform.d/plugins/registry.terraform.io/aiven-dev/aiven/0.0.0+dev/$(ARCH)
BUILD_DEV_BIN ?= $(BUILD_DEV_DIR)/terraform-provider-aiven_v0.0.0+dev

$(BUILD_DEV_DIR):
	mkdir -p $(BUILD_DEV_DIR)

#################################################
# Tools
#################################################
GOLANGCILINT := $(GOTOOL) golangci-lint
TFPLUGINDOCS := $(GOTOOL) tfplugindocs
TERRAFMT := $(GOTOOL) terrafmt
SELPROJ := $(GOTOOL) selproj
MOCKERY := $(GOTOOL) mockery

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
TEST_COUNT ?= 1
ACC_TEST_TIMEOUT ?= 180m
ACC_TEST_PARALLELISM ?= 10

test-acc:
	TF_ACC=1 PROVIDER_AIVEN_ENABLE_BETA=1 $(GO) test ./$(PKG_PATH)/... \
	-v -count $(TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACC_TEST_TIMEOUT)


test-examples: build-dev clean-examples
	AIVEN_PROVIDER_PATH=$(BUILD_DEV_DIR) $(GO) test --tags=examples ./examples_tests/... \
	-v -count $(TEST_COUNT) -parallel $(ACC_TEST_PARALLELISM) $(RUNARGS) $(TESTARGS) -timeout $(ACC_TEST_TIMEOUT)

#################################################
# Lint
#################################################

lint: lint-go lint-test lint-docs semgrep

lint-go:
	$(GOLANGCILINT) run --build-tags all --timeout=30m ./...

lint-test:
	$(TERRAFMT) diff ./internal -cfq

lint-docs:
	PROVIDER_AIVEN_ENABLE_BETA=1 $(TFPLUGINDOCS) generate --rendered-website-dir tmp
	mv tmp/data-sources/influxdb*.md docs/data-sources/
	mv tmp/resources/influxdb*.md docs/resources/
	rm -rf tmp
	PROVIDER_AIVEN_ENABLE_BETA=1 $(TFPLUGINDOCS) validate --provider-name aiven
	rm -f docs/data-sources/influxdb*.md
	rm -f docs/resources/influxdb*.md

semgrep:
	$(CONTAINER_TOOL) run --rm -v "${PWD}:/src" semgrep/semgrep semgrep \
		--config="p/auto" \
		--config=".semgrep.yml" \
		--include="**" \
		--metrics=off \
		--error

#################################################
# Format
#################################################

fmt: fmt-test fmt-imports

fmt-test:
	$(TERRAFMT) fmt ./internal -fv


# macOS requires to install GNU sed first. Use `brew install gnu-sed` to install it.
# It has to be added to PATH as `sed` command, to replace default BSD sed.
# See `brew info gnu-sed` for more details on how to add it to PATH.
# /^import ($$/: starts with "import ("
# /^)/: ends with ")"
# /^[[:space:]]*$$/: empty lines
fmt-imports:
	find . -type f -name '*.go' -exec sed -i '/^import ($$/,/^)/ {/^[[:space:]]*$$/d}' {} +
	goimports -local "github.com/aiven/terraform-provider-aiven" -w .

#################################################
# Clean
#################################################

clean: clean-examples sweep

clean-examples:
	find ./examples -type f -name '*.tfstate*' -delete

SWEEP ?= global

sweep:
	@echo 'WARNING: This will destroy infrastructure. Use only in development accounts.'
	TF_SWEEP=1 $(GO) test ./internal/sweep -v -sweep=$(SWEEP) $(SWEEP_ARGS) -timeout 15m

sweep-check:
	$(GO) test ./internal/sweep -v -run TestCheckSweepers

#################################################
# Generate
#################################################

generate: gen-go docs

gen-go:
	go generate ./...;
	$(MAKE) fmt-imports

docs:
	rm -f docs/.DS_Store
	PROVIDER_AIVEN_ENABLE_BETA=1 $(TFPLUGINDOCS) generate
	rm -f docs/data-sources/influxdb*.md
	rm -f docs/resources/influxdb*.md

OLD_SCHEMA ?= .oldSchema.json
CHANGELOG := PROVIDER_AIVEN_ENABLE_BETA=1 go run ./changelog/...
dump-schemas:
	$(CHANGELOG) -save -schema=$(OLD_SCHEMA)

diff-schemas:
	$(CHANGELOG) -diff -schema=$(OLD_SCHEMA) -changelog=CHANGELOG.md
	rm $(OLD_SCHEMA)

load-schemas:
	go get github.com/aiven/go-client-codegen@latest github.com/aiven/go-api-schemas@latest
	go mod tidy

mockery:
	$(MOCKERY) --config=./.mockery.yml
	$(MAKE) fmt-imports

update-schemas: dump-schemas load-schemas generate diff-schemas mockery

#################################################
# CI
#################################################

ci-selproj:
	$(SELPROJ)