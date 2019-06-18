version = $(shell git describe --long --tags 2>/dev/null || echo unknown-g`git describe --always`)

.PHONY: ci
ci: lint bins

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

bins: vendor
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 packr build -mod=vendor -ldflags "-X main.version=${version}" -o terraform-provider-aiven-linux_amd64 .
	GOOS=darwin GOARCH=amd64 packr build -mod=vendor -ldflags "-X main.version=${version}" -o terraform-provider-aiven-darwin_amd64 .

#################################################
# Testing and linting
#################################################

test: vendor
	CGO_ENABLED=0 go test -v ./...

lint: vendor
	golangci-lint run -D errcheck

clean:
	rm -rf vendor
	rm -f terraform-provider-aiven-*_amd64

.PHONY: test lint vendor bootstrap
