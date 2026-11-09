// Package awsvpcpeeringconnection implements AWS project VPC peering.
//
// The project API identifies a connection by its parent VPC, peer account, peer VPC and region.
// The SDKv2 resource stored these as `PROJECT/PROJECT_VPC_ID/AWS_ACCOUNT_ID/AWS_VPC_ID/AWS_VPC_REGION`,
// but also accepted imports without the region. Existing IDs remain unchanged; resource identity
// records the full key separately. Read and Delete resolve legacy IDs into that same key, using
// the saved region when the ID omits it. During replacement, this identifies the old instance
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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.RefreshStateCheck = pluginvpc.PeeringCreateStateCheck
	ResourceOptions.ImportState = importState
	DataSourceOptions.Read = datasourceReadView
}

func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	// Normalize vpc_id and project/project_vpc_id so switching input forms for the same VPC doesn't replace the connection.
	return pluginvpc.PlanProjectVPC(d)
}

func connectionID(d adapter.ResourceData) (*pluginvpc.ProjectPeeringID, error) {
	if value := d.ID(); value != "" {
		id, err := pluginvpc.ParseProjectPeeringID(value)
		if err != nil {
			return nil, err
		}
		if id.PeerRegion == nil {
			if region, ok := d.GetStateOk("aws_vpc_region"); ok && region.(string) != "" {
				id.PeerRegion = new(region.(string))
			}
		}
		return id, nil
	}
	// Creation and identity imports already supply the individual key fields.
	return &pluginvpc.ProjectPeeringID{
		Project: d.Get("project").(string), ProjectVPCID: d.Get("project_vpc_id").(string),
		PeerCloudAccount: d.Get("aws_account_id").(string), PeerVPC: d.Get("aws_vpc_id").(string),
		PeerRegion: new(d.Get("aws_vpc_region").(string)),
	}, nil
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := connectionID(d)
	if err != nil {
		return err
	}
	if err := id.Create(ctx, client, &vpc.VpcPeeringConnectionCreateIn{
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
	id, err := connectionID(d)
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
	previousID := d.ID()
	err := d.Flatten(connection,
		adapter.RenameFields(map[string]string{
			"peer_cloud_account": "aws_account_id",
			"peer_vpc":           "aws_vpc_id",
			"peer_region":        "aws_vpc_region",
		}),
		pluginvpc.ProjectPeeringFields(ctx, id, "AWS"),
		func(_ adapter.ResourceData, dto map[string]any) error {
			// The API can omit the AWS pcx ID after rejection. Clear any stale value;
			// Flatten normalizes an empty string to Terraform null.
			dto["aws_vpc_peering_connection_id"] = ""
			if cloudID, ok := connection.StateInfo["aws_vpc_peering_connection_id"].(string); ok {
				dto["aws_vpc_peering_connection_id"] = cloudID
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	// Flatten builds a full ID from the identity fields. Preserve the existing
	// string so adding resource identity doesn't change references to legacy IDs.
	if previousID != "" && d.IsResource() {
		return d.SetID(previousID)
	}
	return nil
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := connectionID(d)
	if err != nil {
		return err
	}
	region, err := id.ResolveRegion(ctx, client)
	if err != nil {
		return err
	}
	if id.PeerRegion != nil && region == "" {
		// Only use the regionless route if an exact regionless match still exists.
		if _, err := id.Find(ctx, client); err != nil {
			return err
		}
	}
	id.PeerRegion = util.NilIfZero(region)
	return id.Delete(ctx, client, nil)
}

func importState(ctx context.Context, client avngen.Client, req resource.ImportStateRequest, rsp *resource.ImportStateResponse) {
	id, err := pluginvpc.ParseProjectPeeringID(req.ID)
	if err != nil {
		rsp.Diagnostics.AddError("Invalid peering ID", err.Error())
		return
	}
	region, err := id.ResolveRegion(ctx, client)
	if err != nil {
		rsp.Diagnostics.AddError("Unable to resolve peering region", err.Error())
		return
	}
	for name, value := range map[string]string{
		"id": req.ID, "project": id.Project, "project_vpc_id": id.ProjectVPCID,
		"vpc_id":         schemautil.BuildResourceID(id.Project, id.ProjectVPCID),
		"aws_account_id": id.PeerCloudAccount, "aws_vpc_id": id.PeerVPC, "aws_vpc_region": region,
	} {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(name), value)...)
	}
}
