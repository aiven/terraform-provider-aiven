package main

import (
	"context"
	"flag"
	"log"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")

	flag.Parse()

	ctx := context.Background()

	sdkProvider, err := tf5to6server.UpgradeServer(context.Background(), sdkprovider.Provider().GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Add other provider servers here.
	//  For now, we are wrapping the old Plugin SDK (protocol version 5) server in a new Plugin Framework
	//  (protocol version 6) server, without any changes to the provider itself.
	//  When we are ready to introduce new resources, we should create a new provider server using the new Plugin
	//  Framework (protocol version 6), and add it to the list of providers here.
	//  We cannot make new resources with the old Plugin SDK (protocol version 5) anymore, because the new Plugin
	//  Framework (protocol version 6) is the preferred way to write Terraform providers.
	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return sdkProvider
		},
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
