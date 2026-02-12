package application

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

// planModifier resolves application name to application_id for the data source.
// When name is provided instead of application_id, it lists all applications and finds the matching one.
func planModifier(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	if state.ApplicationID.ValueString() != "" {
		return nil
	}

	var diags diag.Diagnostics

	apps, err := client.ServiceFlinkListApplications(ctx, state.Project.ValueString(), state.ServiceName.ValueString())
	if err != nil {
		diags.Append(errmsg.FromError("ServiceFlinkListApplications Error", err))
		return diags
	}

	for _, app := range apps {
		if app.Name == state.Name.ValueString() {
			state.ApplicationID = types.StringValue(app.Id)
			state.SetID(state.Project.ValueString(), state.ServiceName.ValueString(), app.Id)
			return nil
		}
	}

	diags.AddError(
		"Not Found",
		fmt.Sprintf("flink application %q not found in project %q, service %q",
			state.Name.ValueString(), state.Project.ValueString(), state.ServiceName.ValueString()),
	)
	return diags
}
