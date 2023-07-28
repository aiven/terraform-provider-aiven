package main

import (
	"context"
	"flag"
	"log"

	frameworkprovider "github.com/aiven/terraform-provider-aiven/internal/provider"
	sdkprovider "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

//go:generate go test -tags userconfig ./internal/schemautil/userconfig

// version is the version of the provider.
var version = "dev"

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")

	flag.Parse()

	ctx := context.Background()

	sdkProvider, err := tf5to6server.UpgradeServer(context.Background(), sdkprovider.Provider(version).GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return sdkProvider
		},
		providerserver.NewProtocol6(frameworkprovider.New(version)()),
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt

	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err = tf6server.Serve(
		"registry.terraform.io/aiven/aiven",
		muxServer.ProviderServer,
		serveOpts...,
	); err != nil {
		log.Fatal(err)
	}
}
