package aiven

import (
	"context"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenFlinkJobSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the Flink Job to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the Flink Job to",
		Required:    true,
		ForceNew:    true,
	},
	"job_name": {
		Type:        schema.TypeString,
		Description: "Name of the Flink Job",
		Required:    true,
		ForceNew:    true,
	},
	"job_id": {
		Type:        schema.TypeString,
		Description: "Id of the Flink Job",
		Computed:    true,
	},
	"statement": {
		Type:        schema.TypeString,
		Description: "The SQL Statement of the Flink Job",
		Required:    true,
		ForceNew:    true,
	},
	"tables": {
		Type:        schema.TypeList,
		Description: "The list of tables required in the Job runtime",
		Required:    true,
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"state": {
		Type:        schema.TypeString,
		Description: "The current state of the flink job",
		Computed:    true,
	},
}

func resourceFlinkJob() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceFlinkJobRead,
		CreateContext: resourceFlinkJobCreate,
		DeleteContext: resourceFlinkJobDelete,
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: aivenFlinkJobSchema,
	}
}

func resourceFlinkJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, jobId := splitResourceID3(d.Id())

	r, err := client.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobId})
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
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
	jobTables := flattenToString(d.Get("tables").([]interface{}))

	createRequest := aiven.CreateFlinkJobRequest{
		JobName:   jobName,
		Statement: jobStatement,
		Tables:    jobTables,
	}

	createResponse, err := client.FlinkJobs.Create(project, serviceName, createRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	jobId := createResponse.JobId

	conf := &resource.StateChangeConf{
		Pending: []string{
			"CANCELED",
			"CANCELING",
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
			r, err := client.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobId})
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

	d.SetId(buildResourceID(project, serviceName, r.(*aiven.GetFlinkJobResponse).JID))

	return resourceFlinkJobRead(ctx, d, m)
}

func resourceFlinkJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, jobId := splitResourceID3(d.Id())

	err := client.FlinkJobs.Patch(
		project,
		serviceName,
		aiven.PatchFlinkJobRequest{JobId: jobId},
	)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting flink job: %s", err)
	}

	conf := &resource.StateChangeConf{
		Pending: []string{
			"CANCELING",
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
			r, err := client.FlinkJobs.Get(project, serviceName, aiven.GetFlinkJobRequest{JobId: jobId})
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
