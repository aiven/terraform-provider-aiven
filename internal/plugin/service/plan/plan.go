package plan

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	avivenproject "github.com/aiven/go-client-codegen/handler/project"
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
		Read:     readPlan,
	})
}

func readPlan(ctx context.Context, client avngen.Client, state *tfModel) diag.Diagnostics {
	var (
		diags diag.Diagnostics

		project     = state.Project.ValueString()
		serviceType = state.ServiceType.ValueString()
		servicePlan = state.ServicePlan.ValueString()
		cloudName   = state.CloudName.ValueString()
	)

	specsRsp, err := client.ProjectServicePlanSpecsGet(ctx, project, serviceType, servicePlan)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, fmt.Sprintf("Failed to get plan specs: %s", err.Error()))
		return diags
	}

	priceRsp, err := client.ProjectServicePlanPriceGet(ctx, project, serviceType, servicePlan, cloudName)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, fmt.Sprintf("Failed to get plan pricing: %s", err.Error()))
		return diags
	}

	mergePricing := func(rsp util.RawMap, _ *avivenproject.ProjectServicePlanSpecsGetOut) error {
		if err = rsp.Set(priceRsp.BasePriceUsd, "base_price_usd"); err != nil {
			return fmt.Errorf("failed to set base_price_usd: %w", err)
		}

		if priceRsp.ObjectStorageGbPriceUsd != nil {
			if err = rsp.Set(*priceRsp.ObjectStorageGbPriceUsd, "object_storage_gb_price_usd"); err != nil {
				return fmt.Errorf("failed to set object_storage_gb_price_usd: %w", err)
			}
		}

		if err = rsp.Set(cloudName, "cloud_name"); err != nil {
			return fmt.Errorf("failed to set cloud_name: %w", err)
		}

		if err = rsp.Set(servicePlan, "service_plan"); err != nil {
			return fmt.Errorf("failed to set service_plan: %w", err)
		}

		return nil
	}

	return flattenData(ctx, state, specsRsp, mergePricing)
}
