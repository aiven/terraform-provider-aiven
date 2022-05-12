package main

import (
	"flag"

	"github.com/aiven/terraform-provider-aiven/internal/provider"

	plugin "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate ./internal/schemautil/templates/gen.sh service service_user_config_schema.json
//go:generate ./internal/schemautil/templates/gen.sh integration integrations_user_config_schema.json
//go:generate ./internal/schemautil/templates/gen.sh endpoint integration_endpoints_user_config_schema.json

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
