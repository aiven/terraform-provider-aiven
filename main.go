package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/aiven/terraform-provider-aiven/aiven"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aiven.Provider})
}
