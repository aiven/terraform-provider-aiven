package azurevpcpeeringconnection

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

func init() {
	ResourceOptions.RefreshStateCheck = pluginvpc.PeeringCreateStateCheck
	DataSourceOptions.Read = datasourceReadView
}

func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	return pluginvpc.PlanProjectVPC(d)
}

func configuredID(d adapter.ResourceData) (*pluginvpc.ProjectPeeringID, error) {
	return pluginvpc.NewProjectPeeringID(d, "azure_subscription_id", "vnet_name")
}

func matchResourceGroup(group string) func(*vpc.PeeringConnectionOut) bool {
	return func(connection *vpc.PeeringConnectionOut) bool {
		return group == "" || connection.PeerResourceGroup == group
	}
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := configuredID(d)
	if err != nil {
		return err
	}
	err = id.Create(ctx, client, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount:  id.PeerCloudAccount,
		PeerVpc:           id.PeerVPC,
		PeerResourceGroup: util.NilIfZero(d.Get("peer_resource_group").(string)),
		PeerAzureAppId:    util.NilIfZero(d.Get("peer_azure_app_id").(string)),
		PeerAzureTenantId: util.NilIfZero(d.Get("peer_azure_tenant_id").(string)),
	}, false)
	if err != nil {
		return err
	}
	return d.SetID(id.String())
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := pluginvpc.ParseProjectPeeringID(d.ID())
	if err != nil {
		return err
	}
	connection, err := id.FindMatching(ctx, client, matchResourceGroup(d.Get("peer_resource_group").(string)))
	if err != nil {
		return err
	}
	return setConnectionState(ctx, d, id, connection)
}

func datasourceReadView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := configuredID(d)
	if err != nil {
		return err
	}
	matchGroup := matchResourceGroup(d.Get("peer_resource_group").(string))
	appID := d.Get("peer_azure_app_id").(string)
	tenantID := d.Get("peer_azure_tenant_id").(string)
	connection, err := id.FindMatching(ctx, client, func(connection *vpc.PeeringConnectionOut) bool {
		return matchGroup(connection) && connection.PeerAzureAppId == appID && connection.PeerAzureTenantId == tenantID
	})
	if err != nil {
		return err
	}
	return setConnectionState(ctx, d, id, connection)
}

func setConnectionState(ctx context.Context, d adapter.ResourceData, id *pluginvpc.ProjectPeeringID, connection *vpc.PeeringConnectionOut) error {
	id.PeerRegion = nil
	if err := d.SetID(id.String()); err != nil {
		return err
	}
	return d.Flatten(connection,
		adapter.RenameFields(map[string]string{"peer_cloud_account": "azure_subscription_id", "peer_vpc": "vnet_name"}),
		pluginvpc.ProjectPeeringFields(ctx, id, "Azure"),
		func(_ adapter.ResourceData, dto map[string]any) error {
			// Project VpcGet has no top-level peering ID. Preserve this legacy field;
			// an organization-scoped Aiven ID would not be an Azure cloud ID either.
			delete(dto, "peering_connection_id")
			return nil
		},
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := pluginvpc.ParseProjectPeeringID(d.ID())
	if err != nil {
		return err
	}
	group := d.Get("peer_resource_group").(string)
	if group == "" {
		// Resource group is absent from legacy IDs. Resolve it for an ID-only state.
		connection, err := id.Find(ctx, client)
		if err != nil {
			return err
		}
		group = connection.PeerResourceGroup
	}
	if group == "" {
		return fmt.Errorf("azure VPC peering connection %q has no resource group", id)
	}
	return id.Delete(ctx, client, &group)
}
