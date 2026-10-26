package gcpvpcpeeringconnection

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
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
	return pluginvpc.NewProjectPeeringID(d, "gcp_project_id", "peer_vpc")
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := configuredID(d)
	if err != nil {
		return err
	}

	err = id.Create(ctx, client, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: id.PeerCloudAccount,
		PeerVpc:          id.PeerVPC,
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
	connection, err := id.Find(ctx, client)
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
	connection, err := id.Find(ctx, client)
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
		adapter.RenameFields(map[string]string{"peer_cloud_account": "gcp_project_id"}),
		pluginvpc.ProjectPeeringFields(ctx, id, "Google Cloud"),
		func(d adapter.ResourceData, dto map[string]any) error {
			// The network survives peering failures that omit its details from state_info.
			dto["self_link"], _ = d.GetStateOk("self_link")
			project, _ := connection.StateInfo["to_project_id"].(string)
			network, _ := connection.StateInfo["to_vpc_network"].(string)
			if project != "" && network != "" {
				dto["self_link"] = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, network)
			}
			return nil
		},
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := pluginvpc.ParseProjectPeeringID(d.ID())
	if err != nil {
		return err
	}
	return id.Delete(ctx, client, nil)
}
