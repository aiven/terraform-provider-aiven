ci: lint bins
.PHONY: ci

#################################################
# Bootstrapping for base golang package deps
#################################################

bootstrap:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $$(go env GOPATH)/bin
	go get github.com/gobuffalo/packr/...

vendor:
	go mod tidy
	go mod vendor

update-vendor:


#################################################
# Building
#################################################

bins: vendor
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 packr build -mod=vendor -o terraform-provider-aiven-linux_amd64 .
	GOOS=darwin GOARCH=amd64 packr build -mod=vendor -o terraform-provider-aiven-darwin_amd64 .

#################################################
# Testing and linting
#################################################

test: vendor
	CGO_ENABLED=0 go test -v ./...

lint:
	golangci-lint run -D errcheck

clean:
	go mod tidy
	rm -rf vendor
	rm -f terraform-provided-aiven-*_amd64

.PHONY: test lint vendor bootstrap
