package main

import (
	"github.com/aiven/terraform-provider-aiven/aiven"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate ./aiven/templates/gen.sh service service_user_config_schema.json
//go:generate ./aiven/templates/gen.sh integration integrations_user_config_schema.json
//go:generate ./aiven/templates/gen.sh endpoint integration_endpoints_user_config_schema.json

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aiven.Provider})
}
