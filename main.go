package main

import (
	"flag"

	"github.com/aiven/terraform-provider-aiven/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate go test -tags user_config ./internal/schemautil/user_config

func main() {
	var (
		debugMode bool
	)

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: provider.Provider}

	if debugMode {
		opts.Debug = true
	}

	plugin.Serve(opts)
}
