package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	sdk "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

func NewMuxServer(ctx context.Context, version string) (tfprotov6.ProviderServer, error) {
	p, err := sdk.Provider(version)
	if err != nil {
		return nil, err
	}

	sdkProvider, err := tf5to6server.UpgradeServer(ctx, p.GRPCProvider)
	if err != nil {
		return nil, err
	}

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return sdkProvider
		},
		providerserver.NewProtocol6(plugin.New(version)()),
	}

	server, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}
	return server.ProviderServer(), nil
}
