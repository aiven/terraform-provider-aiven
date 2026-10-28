package flink

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util/serviceutil"
)

var svc = serviceutil.Service{Type: "flink"}

var (
	expandModifier  = svc.ExpandModifier
	flattenModifier = svc.FlattenModifier
	modifyPlan      = svc.ModifyPlan
)

func init() {
	svc.Wrap(&ResourceOptions, &DataSourceOptions)
	del := ResourceOptions.Delete
	ResourceOptions.Delete = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		// Refuses to delete the running applications, e.g., a deployment in "RUNNING" state
		project, name := d.Get("project").(string), d.Get("service_name").(string)
		apps, err := client.ServiceFlinkListApplications(ctx, project, name)
		if common.IsCritical(err) {
			return fmt.Errorf("error listing Flink service applications: %w", err)
		}

		for _, app := range apps {
			deployments, err := client.ServiceFlinkListApplicationDeployments(ctx, project, name, app.Id)
			if common.IsCritical(err) {
				return fmt.Errorf("error listing Flink service deployments: %w", err)
			}

			for _, v := range deployments {
				if v.Status != "CANCELED" {
					return fmt.Errorf(
						"cannot delete Flink service while there are running deployments: %s in state: %s; "+
							"please delete the deployment first and try again", v.Id, v.Status)
				}
			}
		}
		return del(ctx, client, d)
	}
}
