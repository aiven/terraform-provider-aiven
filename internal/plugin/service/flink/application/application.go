package application

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

func init() {
	DataSourceOptions.Read = datasourceReadView
	DataSourceOptions.Schema = datasourceSchemaByName
}

// datasourceSchemaByName wraps the generated datasource schema to require name instead of application_id for lookups.
func datasourceSchemaByName(ctx context.Context) schema.Schema {
	s := datasourceSchema(ctx)
	s.Attributes["name"] = schema.StringAttribute{
		MarkdownDescription: "Application name.",
		Required:            true,
		Validators:          []validator.String{stringvalidator.LengthAtMost(128)},
	}
	s.Attributes["application_id"] = schema.StringAttribute{
		MarkdownDescription: "Application ID.",
		Computed:            true,
	}
	return s
}

// datasourceReadView lists all Flink applications and finds the one matching by name.
func datasourceReadView(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	apps, err := client.ServiceFlinkListApplications(ctx, state.Project.ValueString(), state.ServiceName.ValueString())
	if err != nil {
		diags.Append(errmsg.FromError("ServiceFlinkListApplications Error", err))
		return diags
	}

	for _, app := range apps {
		if app.Name == state.Name.ValueString() {
			rsp, err := client.ServiceFlinkGetApplication(ctx, state.Project.ValueString(), state.ServiceName.ValueString(), app.Id)
			if err != nil {
				diags.Append(errmsg.FromError("ServiceFlinkGetApplication Error", err))
				return diags
			}
			diags.Append(flattenData(ctx, state, rsp)...)
			return diags
		}
	}

	diags.AddError(
		"Not Found",
		fmt.Sprintf("flink application %q not found in project %q, service %q",
			state.Name.ValueString(), state.Project.ValueString(), state.ServiceName.ValueString()),
	)
	return diags
}
