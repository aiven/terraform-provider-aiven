package main

import (
	"context"
	"flag"
	"log"

	"github.com/aiven/terraform-provider-aiven/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate ./aiven/templates/gen.sh service service_user_config_schema.json
//go:generate ./aiven/templates/gen.sh integration integrations_user_config_schema.json
//go:generate ./aiven/templates/gen.sh endpoint integration_endpoints_user_config_schema.json

func main() {
	var (
		debugMode bool
	)

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: provider.Provider}

	if debugMode {
		if err := plugin.Debug(context.Background(), "registry.terraform.io/aiven/aiven", opts); err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}
