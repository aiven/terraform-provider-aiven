package awsprovision

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

// NewResource overrides the generated CRUD:
// - Create: CustomCloudEnvironmentProvision returns (*ProvisionOut, error), generator can't handle 2 return values
// - Delete: no-op, entity lifecycle managed by aiven_byoc_aws_entity
func NewResource() resource.Resource {
	opts := ResourceOptions
	opts.Create = createProvision
	opts.Delete = deleteProvision
	opts.RefreshState = true
	return adapter.NewResource(opts)
}

// createView is a stub referenced by the generated ResourceOptions.
// The YAML has disableView: true for Create, so the generator doesn't produce this function,
// but ResourceOptions still references it. It delegates to createProvision.
func createView(ctx context.Context, client avngen.Client, plan, config *tfModel) diag.Diagnostics {
	return createProvision(ctx, client, plan, config)
}

func createProvision(ctx context.Context, client avngen.Client, plan, _ *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	var req byoc.CustomCloudEnvironmentProvisionIn
	diags.Append(expandData(ctx, plan, nil, &req)...)
	if diags.HasError() {
		return diags
	}

	// CustomCloudEnvironmentProvision returns (*ProvisionOut, error).
	// The generated createView incorrectly discards the response.
	_, err := client.CustomCloudEnvironmentProvision(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.CustomCloudEnvironmentID.ValueString(),
		&req,
	)
	if err != nil {
		diags.Append(errmsg.FromError("CustomCloudEnvironmentProvision Error", err))
		return diags
	}

	return diags
}

func deleteProvision(_ context.Context, _ avngen.Client, _ *tfModel) diag.Diagnostics {
	// No-op: the entity lifecycle is managed by aiven_byoc_aws_entity.
	// Destroying this resource does not deprovision or delete the entity.
	return nil
}
