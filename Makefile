version = $(shell git describe --long --tags 2>/dev/null || echo unknown-g`git describe --always`)
short_version = $(shell echo $(version) | sed 's/-.*//')

.PHONY: ci
ci: lint bins release

#################################################
# Bootstrapping for base golang package deps
#################################################

bootstrap:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.20.0;

#################################################
# Building
#################################################

.PHONY: plugins
plugins:
	mkdir -p plugins/linux_amd64 plugins/darwin_amd64

.PHONY: bins
bins: plugins
	go generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o plugins/linux_amd64/terraform-provider-aiven_$(short_version) .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o plugins/darwin_amd64/terraform-provider-aiven_$(short_version) .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o plugins/windows_amd64/terraform-provider-aiven_$(short_version).exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-X main.version=${version}" -o plugins/windows_386/terraform-provider-aiven_$(short_version).exe .

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

test:
	CGO_ENABLED=0 go test -v --cover ./...

testacc:
	TF_ACC=1 CGO_ENABLED=0 go test -v -count 1 -parallel 10 --cover ./... $(TESTARGS) -timeout 120m

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test -v ./aiven -sweep=global -timeout 60m

lint:
	if [ -z "$(SKIPDIRS)" ]; then \
		golangci-lint run  -D errcheck -D unused -E gofmt --no-config --issues-exit-code=0 --timeout=30m ./...; \
	else \
		golangci-lint run -D errcheck --skip-dirs $(SKIPDIRS) -D unused -E gofmt --no-config --issues-exit-code=0 --timeout=30m ./...; \
	fi

clean:
	rm -rf vendor
	rm -rf plugins
	rm -f terraform-provider-aiven.tar.gz

.PHONY: test lint bootstrap