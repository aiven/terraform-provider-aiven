package planlist

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func NewDataSource() datasource.DataSource {
	return adapter.NewDataSource(adapter.DataSourceOptions[*datasourceModel, tfModel]{
		TypeName: typeName,
		Schema:   datasourceSchema,
		Read:     readPlanList,
	})
}

// planList wraps the API response.
type planList struct {
	Plans []project.ServicePlanOut `json:"service_plans"`
}

func readPlanList(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	rsp, err := client.ProjectServicePlanList(ctx, state.Project.ValueString(), state.ServiceType.ValueString())
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return diags
	}

	list := planList{Plans: rsp}
	return flattenData(ctx, state, &list, regionsToArray)
}

// regionsToArray transforms regions from map[string]any to cloud_names []string.
// We only need the region names for the Terraform schema.
func regionsToArray(rsp util.RawMap, in *planList) error {
	if len(in.Plans) == 0 {
		return nil
	}

	for i, plan := range in.Plans {
		if len(plan.Regions) == 0 {
			continue
		}

		regionNames := make([]string, 0, len(plan.Regions))
		for rn := range plan.Regions {
			regionNames = append(regionNames, rn)
		}

		index := fmt.Sprintf("[%d]", i)
		err := rsp.Set(regionNames, "service_plans", index, "cloud_names")
		if err != nil {
			return fmt.Errorf("failed to set cloud_names for plan %d: %w", i, err)
		}
	}

	return nil
}
