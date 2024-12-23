// Package flink is the package that contains the schema definitions for the Flink resources.
package flink

import (
	"context"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// aivenFlinkApplicationDeploymentSchema is the schema for the Flink Application Deployment resource.
var aivenFlinkApplicationDeploymentSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"application_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Application ID.",
	},
	// Request fields.
	"parallelism": {
		Type:         schema.TypeInt,
		Optional:     true,
		Description:  "The number of parallel instances for the task.",
		ValidateFunc: validation.IntBetween(1, 128),
		ForceNew:     true,
		Default:      1,
	},
	"restart_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Restart a Flink job if it fails.",
		ForceNew:    true,
		Default:     true,
	},
	"starting_savepoint": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The savepoint to deploy from.",
		ValidateFunc: validation.StringLenBetween(1, 2048),
		ForceNew:     true,
	},
	"version_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Application version ID.",
		ForceNew:    true,
	},
	// Computed fields.
	"created_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Application deployment creation time.",
	},
	"created_by": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The user who deployed the application.",
	},
}

// ResourceFlinkApplicationDeployment returns the schema for the Flink Application Deployment resource.
func ResourceFlinkApplicationDeployment() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages the deployment of an Aiven for Apache Flink® application.",
		CreateContext: resourceFlinkApplicationDeploymentCreate,
		ReadContext:   resourceFlinkApplicationDeploymentRead,
		DeleteContext: resourceFlinkApplicationDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenFlinkApplicationDeploymentSchema,
	}
}

// resourceFlinkApplicationDeploymentCreate creates a new Flink Application Deployment resource.
func resourceFlinkApplicationDeploymentCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationID := d.Get("application_id").(string)

	var req aiven.CreateFlinkApplicationDeploymentRequest

	if v, ok := d.GetOk("parallelism"); ok {
		req.Parallelism = v.(int)
	}

	if v, ok := d.GetOk("restart_enabled"); ok {
		req.RestartEnabled = v.(bool)
	}

	if v, ok := d.GetOk("starting_savepoint"); ok {
		req.StartingSavepoint = v.(string)
	}

	if v, ok := d.GetOk("version_id"); ok {
		req.VersionID = v.(string)
	}

	r, err := client.FlinkApplicationDeployments.Create(ctx, project, serviceName, applicationID, req)
	if err != nil {
		return diag.Errorf("cannot create Flink Application Deployment: %v", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, applicationID, r.ID))

	return resourceFlinkApplicationDeploymentRead(ctx, d, m)
}

// resourceFlinkApplicationDeploymentDelete deletes an existing Flink Application Deployment resource.
func resourceFlinkApplicationDeploymentDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, applicationID, deploymentID, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.Errorf("cannot read Flink Application Deployment resource ID: %v", err)
	}

	// Flink Application Deployment has a quite complicated state machine
	// https://api.aiven.io/doc/#tag/Service:_Flink/operation/ServiceFlinkDeleteApplicationDeployment
	// Retries until succeeds or exceeds the timeout
	for {
		select {
		case <-ctx.Done():
			// The context itself already comes with delete timeout
			return diag.Errorf("can't delete Flink Application Deployment: %s", ctx.Err())
		case <-time.After(time.Second):
			_, err := client.FlinkApplicationDeployments.Get(ctx, project, serviceName, applicationID, deploymentID)
			if aiven.IsNotFound(err) {
				return nil
			}

			// Must be canceled before deleted
			_, err = client.FlinkApplicationDeployments.Cancel(ctx, project, serviceName, applicationID, deploymentID)
			if err == nil {
				continue
			}

			// Completely ignores all errors, until it gets 404 on GET request
			_, _ = client.FlinkApplicationDeployments.Delete(ctx, project, serviceName, applicationID, deploymentID)
		}
	}
}

// resourceFlinkApplicationDeploymentRead reads an existing Flink Application Deployment resource.
func resourceFlinkApplicationDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, applicationID, deploymentID, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.Errorf("cannot read Flink Application Deployment resource ID: %v", err)
	}

	r, err := client.FlinkApplicationDeployments.Get(ctx, project, serviceName, applicationID, deploymentID)
	if err != nil {
		return diag.Errorf("cannot get Flink Application Deployment: %v", err)
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `project` field: %s", err)
	}

	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `service_name` field: %s", err)
	}

	if err := d.Set("application_id", applicationID); err != nil {
		return diag.Errorf("error setting Flink Application Version `application_id` field: %s", err)
	}

	if err := d.Set("parallelism", r.Parallelism); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `parallelism` field: %s", err)
	}

	if err := d.Set("restart_enabled", r.RestartEnabled); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `restart_enabled` field: %s", err)
	}

	if err := d.Set("starting_savepoint", r.StartingSavepoint); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `starting_savepoint` field: %s", err)
	}

	if err := d.Set("version_id", r.VersionID); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `version_id` field: %s", err)
	}

	if err := d.Set("created_at", r.CreatedAt); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `created_at` field: %s", err)
	}

	if err := d.Set("created_by", r.CreatedBy); err != nil {
		return diag.Errorf("error setting Flink Application Deployment `created_by` field: %s", err)
	}

	return nil
}
