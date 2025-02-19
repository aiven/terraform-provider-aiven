package flink

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/flinkjarapplication"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func ResourceFlinkJarApplication() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Apache Flink® jar application](https://aiven.io/docs/products/flink/howto/create-jar-application).",
		ReadContext:   common.WithGenClient(flinkJarApplicationRead),
		CreateContext: common.WithGenClient(flinkJarApplicationCreate),
		UpdateContext: common.WithGenClient(flinkJarApplicationUpdate),
		DeleteContext: common.WithGenClient(flinkJarApplicationDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   flinkJarApplicationSchema(),
	}
}

func flinkJarApplicationCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	req := new(flinkjarapplication.ServiceFlinkCreateJarApplicationIn)
	err := schemautil.ResourceDataGet(d, req, schemautil.RenameAliases(flinkJarApplicationRename()))
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	rsp, err := client.ServiceFlinkCreateJarApplication(ctx, project, serviceName, req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, rsp.Id))
	return flinkJarApplicationRead(ctx, d, client)
}

func flinkJarApplicationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, applicationID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	rsp, err := client.ServiceFlinkGetJarApplication(ctx, projectName, serviceName, applicationID)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	// Deployment is created after the application is created.
	// That is a circular dependency, and it fails with a non-empty plan in tests.
	// Setting an empty object to suppress the diff.
	if rsp.CurrentDeployment == nil {
		rsp.CurrentDeployment = new(flinkjarapplication.CurrentDeploymentOut)
	}

	return schemautil.ResourceDataSet(
		d, rsp, flinkJarApplicationSchema(),
		schemautil.RenameAliasesReverse(flinkJarApplicationRename()),
		schemautil.SetForceNew("project", projectName),
		schemautil.SetForceNew("service_name", serviceName),
	)
}

func flinkJarApplicationUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	req := new(flinkjarapplication.ServiceFlinkUpdateJarApplicationIn)
	err := schemautil.ResourceDataGet(d, req)
	if err != nil {
		return err
	}

	project, serviceName, applicationID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	_, err = client.ServiceFlinkUpdateJarApplication(ctx, project, serviceName, applicationID, req)
	if err != nil {
		return err
	}

	return flinkJarApplicationRead(ctx, d, client)
}

func flinkJarApplicationDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, applicationID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	_, err = client.ServiceFlinkDeleteJarApplication(ctx, project, serviceName, applicationID)
	return schemautil.OmitNotFound(err)
}
