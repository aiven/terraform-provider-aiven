//go:build tools
// +build tools

package main

import (
	_ "github.com/ChimeraCoder/gojson/gojson"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "github.com/katbyte/terrafmt"
)
