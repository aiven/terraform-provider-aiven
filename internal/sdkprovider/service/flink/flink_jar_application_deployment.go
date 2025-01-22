// Package flink is the package that contains the schema definitions for the Flink resources.
package flink

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/flinkjarapplicationdeployment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// ResourceFlinkJarApplicationDeployment returns the schema for the Flink Jar Application Deployment resource.
func ResourceFlinkJarApplicationDeployment() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages the deployment of an Aiven for Apache Flink® application.",
		CreateContext: common.WithGenClient(flinkApplicationDeploymentCreate),
		ReadContext:   common.WithGenClient(flinkApplicationDeploymentRead),
		DeleteContext: common.WithGenClient(flinkApplicationDeploymentDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   flinkJarApplicationDeploymentSchema(),
	}
}

func flinkApplicationDeploymentCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req flinkjarapplicationdeployment.ServiceFlinkCreateJarApplicationDeploymentIn
	err := schemautil.ResourceDataGet(d, &req, schemautil.RenameAliases(flinkJarApplicationDeploymentRename()))
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationID := d.Get("application_id").(string)
	r, err := client.ServiceFlinkCreateJarApplicationDeployment(ctx, project, serviceName, applicationID, &req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, applicationID, r.Id))
	return flinkApplicationDeploymentRead(ctx, d, client)
}

func flinkApplicationDeploymentDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, applicationID, deploymentID, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	// Flink Jar Application Deployment has a quite complicated state machine
	// https://api.aiven.io/doc/#tag/Service:_Flink/operation/ServiceFlinkGetJarApplicationDeployment
	// Retries until succeeds or exceeds the timeout
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		_, err := client.ServiceFlinkGetJarApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		if avngen.IsNotFound(err) {
			return nil
		}

		// Must be canceled before deleted
		_, err = client.ServiceFlinkCancelJarApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		if err != nil {
			// Nothing to cancel.
			// Completely ignores all errors, until it gets 404 on GET request
			_, _ = client.ServiceFlinkDeleteJarApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		}

		select {
		case <-ctx.Done():
			// The context itself already comes with delete timeout
			return fmt.Errorf("can't delete Flink Application Deployment: %w", ctx.Err())
		case <-ticker.C:
			continue
		}
	}
}

func flinkApplicationDeploymentRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, applicationID, deploymentID, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	rsp, err := client.ServiceFlinkGetJarApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
	if err != nil {
		return err
	}

	return schemautil.ResourceDataSet(
		flinkJarApplicationDeploymentSchema(), d, rsp,
		schemautil.RenameAliasesReverse(flinkJarApplicationDeploymentRename()),
	)
}
