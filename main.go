package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/acctest/template"
	"github.com/aiven/terraform-provider-aiven/internal/exporter"
	"github.com/aiven/terraform-provider-aiven/internal/server"
)

//go:generate go run ./ucgenerator/...

// registryPrefix is the registry prefix for the provider.
const registryPrefix = "registry.terraform.io/"

// version is the version of the provider.
var version = "dev"

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")

	flag.Parse()
	ctx := context.Background()
	muxServer, err := server.NewMuxServer(ctx, version)
	if err != nil {
		log.Panic(err)
	}

	var serveOpts []tf6server.ServeOpt
	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	name := registryPrefix + "aiven/aiven"

	//goland:noinspection GoBoolExpressions
	if version == "dev" {
		name = registryPrefix + "aiven-dev/aiven"
	}

	exporter.RegisterExportHook(func(d *schema.ResourceData, resourceType string) error {
		return template.ExportResourceAsHCL(d, resourceType)
	})

	err = tf6server.Serve(
		name,
		func() tfprotov6.ProviderServer {
			return muxServer
		},
		serveOpts...,
	)
	if err != nil {
		log.Panic(err)
	}
}
