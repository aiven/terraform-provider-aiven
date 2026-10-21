package awsvpcpeeringconnection

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

func init() {
	ResourceOptions.RefreshStateCheck = pluginvpc.PeeringRefreshStateCheck
	DataSourceOptions.Read = datasourceReadView
}

// The SDK suppressed every diff to an empty region. Framework must keep that
// configured value, so it is stored by a local update. Reads and deletes then
// resolve the actual region from the API.
func modifyPlan(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if err := pluginvpc.PlanProjectVPC(d); err != nil {
		return err
	}
	if d.IsNewResource() {
		return nil
	}
	region, known := d.GetOk("aws_vpc_region")
	if !known {
		d.RequiresReplace("aws_vpc_region")
		return nil
	}
	if region == "" || !d.HasChange("aws_vpc_region") {
		return nil
	}
	id, err := stateID(d, d.GetState("aws_vpc_region").(string))
	if err != nil {
		return err
	}
	actualRegion, err := resolveRegion(ctx, client, id)
	if err != nil {
		return err
	}
	if region != actualRegion {
		d.RequiresReplace("aws_vpc_region")
	}
	return nil
}

func configuredID(d adapter.ResourceData) (*pluginvpc.ProjectPeeringID, error) {
	id, err := pluginvpc.NewProjectPeeringID(d, "aws_account_id", "aws_vpc_id")
	if err != nil {
		return nil, err
	}
	region := d.Get("aws_vpc_region").(string)
	if region != "" {
		id.PeerRegion = &region
	}
	return id, nil
}

// stateID takes the region from a legacy five-part ID, then from stateRegion.
// It leaves the region unset for `aws_vpc_region = ""`, matching any region.
func stateID(d adapter.ResourceData, stateRegion string) (*pluginvpc.ProjectPeeringID, error) {
	id, err := pluginvpc.StateProjectPeeringID(d, "aws_account_id", "aws_vpc_id")
	if err != nil {
		return nil, err
	}
	if id.PeerRegion == nil && stateRegion != "" {
		id.PeerRegion = &stateRegion
	}
	return id, nil
}

// resolveRegion returns the region of id, looking up the connection when unset.
func resolveRegion(ctx context.Context, client avngen.Client, id *pluginvpc.ProjectPeeringID) (string, error) {
	if id.PeerRegion != nil {
		return *id.PeerRegion, nil
	}
	connection, err := id.Find(ctx, client)
	if err != nil {
		return "", err
	}
	if connection.PeerRegion == nil {
		return "", nil
	}
	return *connection.PeerRegion, nil
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := configuredID(d)
	if err != nil {
		return err
	}
	return id.Create(ctx, client, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: id.PeerCloudAccount,
		PeerVpc:          id.PeerVPC,
		PeerRegion:       id.PeerRegion,
	})
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := stateID(d, d.Get("aws_vpc_region").(string))
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
	// Data sources require an exact region match, including an explicit empty string.
	id.PeerRegion = new(d.Get("aws_vpc_region").(string))
	connection, err := id.Find(ctx, client)
	if err != nil {
		return err
	}
	return setConnectionState(ctx, d, id, connection)
}

func setConnectionState(ctx context.Context, d adapter.ResourceData, id *pluginvpc.ProjectPeeringID, connection *vpc.PeeringConnectionOut) error {
	region, regionSet := d.GetOk("aws_vpc_region")
	return d.Flatten(connection,
		adapter.RenameFields(map[string]string{
			"peer_cloud_account": "aws_account_id", "peer_vpc": "aws_vpc_id", "peer_region": "aws_vpc_region",
		}),
		pluginvpc.ProjectPeeringFields(ctx, id, "AWS"),
		func(_ adapter.ResourceData, dto map[string]any) error {
			if d.IsResource() && regionSet && region == "" {
				// Keep the configured value already in plan/state. Flatten normalizes
				// empty API strings to null, which would violate the configuration.
				delete(dto, "aws_vpc_region")
			}
			dto["aws_vpc_peering_connection_id"] = nil
			if cloudID, ok := connection.StateInfo["aws_vpc_peering_connection_id"].(string); ok {
				dto["aws_vpc_peering_connection_id"] = cloudID
			}
			return nil
		},
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := stateID(d, d.Get("aws_vpc_region").(string))
	if err != nil {
		return err
	}
	region, err := resolveRegion(ctx, client, id)
	if err != nil {
		return err
	}
	if region != "" {
		id.PeerRegion = &region
	}
	return id.Delete(ctx, client, nil)
}
