package flink

import (
	"context"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenFlinkJobSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"job_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Specifies the name of the service that this job is submitted to.").ForceNew().Referenced().Build(),
	},
	"statement": {
		Type:        schema.TypeString,
		Description: schemautil.Complex("The SQL statement to define the job.").ForceNew().Build(),
		Required:    true,
		ForceNew:    true,
	},
	"table_ids": {
		Type:        schema.TypeList,
		Description: schemautil.Complex("A list of table ids that are required in the job runtime.").ForceNew().Referenced().Build(),
		Required:    true,
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"job_id": {
		Type:        schema.TypeString,
		Description: "The Job ID of the flink job in the flink service.",
		Computed:    true,
	},
	"state": {
		Type:        schema.TypeString,
		Description: "The current state of the flink job in the flink service",
		Computed:    true,
	},
}

func ResourceFlinkJob() *schema.Resource {
	return &schema.Resource{
		Description:        "The Flink Job resource allows the creation and management of Aiven Jobs.",
		ReadContext:        resourceFlinkJobRead,
		CreateContext:      resourceFlinkJobCreate,
		DeleteContext:      resourceFlinkJobDelete,
		Timeouts:           schemautil.DefaultResourceTimeouts(),
		Schema:             aivenFlinkJobSchema,
		CustomizeDiff:      customdiff.If(schemautil.ResourceShouldExist, resourceFlinkJobCustomizeDiff),
		DeprecationMessage: schemautil.DeprecationMessage,
	}
}

func resourceFlinkJobRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, jobID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobID})
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	// we model job deletion by canceling the job
	if r.State == "CANCELED" {
		d.SetId("")
		return nil
	}

	// setting fields from the response that are tracked by the schema
	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting Flink Jobs `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting Flink Jobs `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("job_id", r.JID); err != nil {
		return diag.Errorf("error setting Flink Jobs `job_id` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("job_name", r.Name); err != nil {
		return diag.Errorf("error setting Flink Jobs `job_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("state", r.State); err != nil {
		return diag.Errorf("error setting Flink Jobs `state` for resource %s: %s", d.Id(), err)
	}
	// statement and tables cannot be read remotely; but they are immutable, so just dont touch them

	return nil
}

func resourceFlinkJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	jobName := d.Get("job_name").(string)
	jobStatement := d.Get("statement").(string)
	jobTables := schemautil.FlattenToString(d.Get("table_ids").([]interface{}))

	createRequest := aiven.CreateFlinkJobRequest{
		JobName:   jobName,
		Statement: jobStatement,
		TablesIds: jobTables,
	}

	createResponse, err := client.FlinkJobs.Create(project, serviceName, createRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	jobID := createResponse.JobId

	conf := &resource.StateChangeConf{
		Pending: []string{
			"CANCELED",
			"CANCELLING",
			"CREATED",
			"DEPLOYING",
			"FAILED",
			"FINISHED",
			"INITIALIZING",
			"RECONCILING",
			"SCHEDULED",
		},
		Target: []string{
			"RUNNING",
		},
		Refresh: func() (interface{}, string, error) {
			r, err := client.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobID})
			if err != nil {
				return nil, "", err
			}
			return r, r.State, nil
		},
		Delay:      1 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Second,
	}

	r, err := conf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for job to become active: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, r.(*aiven.GetFlinkJobResponse).JID))

	return resourceFlinkJobRead(ctx, d, m)
}

func resourceFlinkJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, jobID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.FlinkJobs.Patch(
		project,
		serviceName,
		aiven.PatchFlinkJobRequest{JobId: jobID},
	)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting flink job: %s", err)
	}

	conf := &resource.StateChangeConf{
		Pending: []string{
			"CANCELLING",
			"CREATED",
			"DEPLOYING",
			"INITIALIZING",
			"RECONCILING",
			"RUNNING",
			"SCHEDULED",
		},
		// flink does not cancel job that have failed or finished
		// so we accept these states also as "deleted", otherwise we will
		// loop endless here
		Target: []string{
			"CANCELED",
			"FAILED",
			"FINISHED",
		},
		Refresh: func() (interface{}, string, error) {
			r, err := client.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobID})
			if err != nil {
				return nil, "", err
			}
			return r, r.State, nil
		},
		Delay:      1 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 1 * time.Second,
	}

	if _, err = conf.WaitForStateContext(ctx); err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}
	return nil
}

func resourceFlinkJobCustomizeDiff(_ context.Context, d *schema.ResourceDiff, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	_, err := client.FlinkJobs.Validate(projectName, serviceName, aiven.ValidateFlinkJobRequest{
		Statement: d.Get("statement").(string),
		TableIDs:  schemautil.FlattenToString(d.Get("table_ids").([]interface{})),
	})

	return err
}
