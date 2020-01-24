version = $(shell git describe --long --tags 2>/dev/null || echo unknown-g`git describe --always`)
short_version = $(shell echo $(version) | sed 's/-.*//')

.PHONY: ci
ci: lint bins release

#################################################
# Bootstrapping for base golang package deps
#################################################

bootstrap:
	if [ -z "$$(which golangci-lint 2>/dev/null)" ]; then \
 	  curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi
	go get github.com/gobuffalo/packr/...

vendor:
	go mod vendor

update-vendor:


#################################################
# Building
#################################################

.PHONY: plugins
plugins:
	mkdir -p plugins/linux_amd64 plugins/darwin_amd64

.PHONY: bins
bins: vendor plugins
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 packr2 build -mod=vendor -ldflags "-X main.version=${version}" -o plugins/linux_amd64/terraform-provider-aiven_$(short_version) .
	GOOS=darwin GOARCH=amd64 packr2 build -mod=vendor -ldflags "-X main.version=${version}" -o plugins/darwin_amd64/terraform-provider-aiven_$(short_version) .
	GOOS=windows GOARCH=amd64 packr2 build -mod=vendor -ldflags "-X main.version=${version}" -o plugins/windows_amd64/terraform-provider-aiven_$(short_version).exe .
	GOOS=windows GOARCH=386 packr2 build -mod=vendor -ldflags "-X main.version=${version}" -o plugins/windows_386/terraform-provider-aiven_$(short_version).exe .

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
# Testing and linting
#################################################

test: vendor
	CGO_ENABLED=0 go test -v ./...

testacc: vendor
	TF_ACC=1 CGO_ENABLED=0 go test -v --cover ./... -timeout 120m

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test -v ./... -sweep=global -timeout 60m

lint: vendor
	golangci-lint run  -D errcheck -D unused -E gofmt --no-config --issues-exit-code=0 --timeout=30m ./...

clean:
	packr2 clean
	rm -rf vendor
	rm -rf plugins
	rm -f terraform-provider-aiven.tar.gz

.PHONY: test lint vendor bootstrap
