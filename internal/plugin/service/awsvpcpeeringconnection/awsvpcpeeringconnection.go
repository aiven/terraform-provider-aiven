// Package awsvpcpeeringconnection implements AWS project VPC peering.
//
// The project API identifies a connection by its parent VPC, peer account, peer VPC and region.
// The SDKv2 resource stored these as `PROJECT/PROJECT_VPC_ID/AWS_ACCOUNT_ID/AWS_VPC_ID/AWS_VPC_REGION`,
// but also accepted imports without the region. Both formats are preserved to keep existing state and planned IDs stable.
// During replacement, the saved identity must still refer to the old connection,
// even when the configuration already describes its replacement.
//
// SDKv2 accepted `aws_vpc_region = ""`. For a new connection, this let the API choose the Aiven VPC's region.
// For an existing connection, changing the configured region to "" left its region unchanged, even across regions.
// State contained the actual region in either case. Framework cannot replace a configured "" with that value in state.
// An explicit, nonempty region is now required for both resources and data sources; changing it requires replacement.
//
// The creation API can reuse an existing connection with the same identifiers. The provider rejects such matches
// so `create_before_destroy` cannot adopt the old connection and then delete it during replacement cleanup.
// Creation and readiness are separate: a successfully created connection stays in Terraform state even if it fails
// to become ready, so it can still be deleted or replaced.
package awsvpcpeeringconnection

import (
	"context"

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
	// Normalize vpc_id and project/project_vpc_id so switching input forms for the same VPC doesn't replace the connection.
	return pluginvpc.PlanProjectVPC(d)
}

// The ID's region is authoritative. Legacy four-part imports retain the actual region in prior state.
func stateID(d adapter.ResourceData) (*pluginvpc.ProjectPeeringID, error) {
	id, err := pluginvpc.ParseProjectPeeringID(d.ID())
	if err != nil {
		return nil, err
	}
	if id.PeerRegion != nil {
		return id, nil
	}
	if region, ok := d.GetStateOk("aws_vpc_region"); ok {
		id.PeerRegion = util.NilIfZero(region.(string))
	}
	return id, nil
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := pluginvpc.NewProjectPeeringID(d, "aws_account_id", "aws_vpc_id")
	if err != nil {
		return err
	}
	id.PeerRegion = new(d.Get("aws_vpc_region").(string))
	if err = id.Create(ctx, client, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: id.PeerCloudAccount,
		PeerVpc:          id.PeerVPC,
		PeerRegion:       id.PeerRegion,
	}, false); err != nil {
		return err
	}
	// Save the connection ID before waiting for readiness so a failed readiness check doesn't leave
	// the connection outside Terraform state.
	return d.SetID(id.String())
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := stateID(d)
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
	id, err := pluginvpc.NewProjectPeeringID(d, "aws_account_id", "aws_vpc_id")
	if err != nil {
		return err
	}
	id.PeerRegion = new(d.Get("aws_vpc_region").(string))
	connection, err := id.Find(ctx, client)
	if err != nil {
		return err
	}
	return setConnectionState(ctx, d, id, connection)
}

func setConnectionState(ctx context.Context, d adapter.ResourceData, id *pluginvpc.ProjectPeeringID, connection *vpc.PeeringConnectionOut) error {
	resourceID := d.ID()
	if resourceID == "" {
		resourceID = id.String()
	}
	return d.Flatten(connection,
		adapter.RenameFields(map[string]string{
			"peer_cloud_account": "aws_account_id",
			"peer_vpc":           "aws_vpc_id",
			"peer_region":        "aws_vpc_region",
		}),
		pluginvpc.ProjectPeeringFields(ctx, id, "AWS"),
		func(_ adapter.ResourceData, dto map[string]any) error {
			// Preserve existing four-part imports and known planned IDs.
			dto["id"] = resourceID
			// The API can omit the AWS pcx ID after rejection. Clear any stale value;
			// Flatten normalizes an empty string to Terraform null.
			dto["aws_vpc_peering_connection_id"] = ""
			if cloudID, ok := connection.StateInfo["aws_vpc_peering_connection_id"].(string); ok {
				dto["aws_vpc_peering_connection_id"] = cloudID
			}
			return nil
		},
	)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := stateID(d)
	if err != nil {
		return err
	}
	region, err := id.ResolveRegion(ctx, client)
	if err != nil {
		return err
	}
	id.PeerRegion = util.NilIfZero(region)
	return id.Delete(ctx, client, nil)
}
