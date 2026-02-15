package permissions

import (
	"context"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/byoc"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

// NewResource overrides the generated schema and CRUD:
// - Schema: marks accounts/project_names as Optional+Computed (not Required+ForceNew)
// - Update: same as Create (Set endpoint is idempotent, full replacement)
// - Delete: sends empty arrays to revoke all permissions
// - ConfigValidators: requires at least one of accounts or project_names
func NewResource() resource.Resource {
	opts := ResourceOptions
	opts.Schema = permissionsSchema
	opts.Update = updatePermissions
	opts.Delete = deletePermissions
	opts.ConfigValidators = configValidators
	return adapter.NewResource(opts)
}

// permissionsSchema overrides the generated schema.
// The generator marks accounts and project_names as Required + ForceNew because
// they appear in the Create request and there is no Update operation.
// In reality they should be Optional (not Required) and not ForceNew,
// because the Set API performs a full replacement and Update calls the same endpoint.
func permissionsSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"accounts": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Account IDs to grant access to this BYOC cloud. Granting access to an account grants access to all projects under it.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.Set{setvalidator.SizeAtMost(1000)},
			},
			"custom_cloud_environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of a custom cloud environment. Changing this property forces recreation of the resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource ID, a composite of `organization_id` and `custom_cloud_environment_id` IDs.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "ID of an organization. Changing this property forces recreation of the resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"project_names": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Project names to grant access to this BYOC cloud.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.Set{setvalidator.SizeAtMost(1000)},
			},
		},
		Blocks:              map[string]schema.Block{"timeouts": timeouts.BlockAll(ctx)},
		MarkdownDescription: "Grants accounts and projects access to a BYOC custom cloud environment.\n\n**This resource is in the beta stage and may change without notice.** Set\nthe `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.",
	}
}

func configValidators(_ context.Context, _ avngen.Client) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("accounts"),
			path.MatchRoot("project_names"),
		),
	}
}

func updatePermissions(ctx context.Context, client avngen.Client, plan, _, _ *tfModel) diag.Diagnostics {
	// Same as Create: the Set endpoint is idempotent (full replacement)
	var diags diag.Diagnostics
	var req byoc.CustomCloudEnvironmentPermissionsSetIn
	diags.Append(expandData(ctx, plan, nil, &req)...)
	if diags.HasError() {
		return diags
	}

	err := client.CustomCloudEnvironmentPermissionsSet(
		ctx,
		plan.OrganizationID.ValueString(),
		plan.CustomCloudEnvironmentID.ValueString(),
		&req,
	)
	if err != nil {
		diags.Append(errmsg.FromError("CustomCloudEnvironmentPermissionsSet Error", err))
		return diags
	}

	return diags
}

func deletePermissions(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	// Revoke all permissions by setting empty arrays
	var diags diag.Diagnostics
	err := client.CustomCloudEnvironmentPermissionsSet(
		ctx,
		state.OrganizationID.ValueString(),
		state.CustomCloudEnvironmentID.ValueString(),
		&byoc.CustomCloudEnvironmentPermissionsSetIn{
			Accounts: []string{},
			Projects: []string{},
		},
	)
	if err != nil {
		diags.Append(errmsg.FromError("CustomCloudEnvironmentPermissionsSet Error", err))
		return diags
	}

	return diags
}
