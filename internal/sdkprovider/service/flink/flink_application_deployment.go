// Package flink is the package that contains the schema definitions for the Flink resources.
package flink

import (
	"context"
	"errors"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/exp/slices"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// Flink Application Deployment states.
//
// Below states are based on https://nightlies.apache.org/flink/flink-docs-master/docs/internals/job_scheduling/.
// See the link for more information about the states and the state transitions.
const (
	// Initial states.
	// flinkDeploymentInitializingState is the state when the Flink Application Deployment is initializing.
	flinkDeploymentInitializingState = "INITIALIZING"
	// flinkDeploymentCreatedState is the state when the Flink Application Deployment is created.
	flinkDeploymentCreatedState = "CREATED"

	// Active running states.
	// flinkDeploymentRunningState is the state when the Flink Application Deployment is running.
	flinkDeploymentRunningState = "RUNNING"
	// flinkRestartingState is the state when the Flink Application Deployment is restarting.
	flinkRestartingState = "RESTARTING"
	// flinkDeploymentSavingState is the state when the Flink Application Deployment is saving.
	flinkDeploymentSavingState = "SAVING"

	// Intermediate states.
	// flinkDeploymentFailingState is the state when the Flink Application Deployment is failing.
	flinkDeploymentFailingState = "FAILING"
	// flinkDeploymentCancellingRequestedState is the state when the Flink Application Deployment is planning to cancel.
	flinkDeploymentCancellingRequestedState = "CANCELLING_REQUESTED"
	// flinkDeploymentCancellingState is the state when the Flink Application Deployment is cancelling.
	flinkDeploymentCancellingState = "CANCELLING"
	// flinkDeploymentSavingAndStopRequestedState is the state when the Flink Application Deployment is planning to
	// save and stop.
	flinkDeploymentSavingAndStopRequestedState = "SAVING_AND_STOP_REQUESTED"
	// flinkDeploymentSavingAndStopState is the state when the Flink Application Deployment is saving and stopping.
	flinkDeploymentSavingAndStopState = "SAVING_AND_STOP"

	// Terminal states.
	// flinkDeploymentFailedState is the state when the Flink Application Deployment has failed.
	flinkDeploymentFailedState = "FAILED"
	// flinkDeploymentCanceledState is the state when the Flink Application Deployment is canceled.
	flinkDeploymentCanceledState = "CANCELED"
	// flinkDeploymentFinishedState is the state when the Flink Application Deployment is finished.
	flinkDeploymentFinishedState = "FINISHED"
	// flinkDeploymentSuspendedState is the state when the Flink Application Deployment is suspended.
	flinkDeploymentSuspendedState = "SUSPENDED"
	// flinkDeploymentDeleteRequestedState is the state when the Flink Application Deployment is planning to delete.
	flinkDeploymentDeleteRequestedState = "DELETE_REQUESTED"
	// flinkDeploymentDeletingState is the state when the Flink Application Deployment is deleting.
	flinkDeploymentDeletingState = "DELETING"
)

// flinkInitialDeploymentStates returns the list of Flink Application Deployment states that are considered initial.
var flinkInitialDeploymentStates = []string{
	flinkDeploymentInitializingState,
	flinkDeploymentCreatedState,
}

// flinkActiveDeploymentStates returns the list of Flink Application Deployment states that are considered active.
var flinkActiveDeploymentStates = []string{
	flinkDeploymentRunningState,
	flinkDeploymentSavingState,
	flinkRestartingState,
}

// flinkInactiveDeploymentStates returns the list of Flink Application Deployment states that are considered inactive.
var flinkInactiveDeploymentStates = []string{
	flinkDeploymentFailingState,
	flinkDeploymentCancellingState,
	flinkDeploymentFailedState,
	flinkDeploymentCanceledState,
	flinkDeploymentFinishedState,
}

// flinkNonDeletableOrCancelableDeploymentStates returns the list of Flink Application Deployment states that are
// considered non-deletable or non-cancelable.
var flinkNonDeletableOrCancelableDeploymentStates = []string{
	flinkDeploymentInitializingState,
	flinkDeploymentFailingState,
	flinkDeploymentCancellingState,
	flinkDeploymentSavingAndStopRequestedState,
	flinkDeploymentSavingAndStopState,
	flinkDeploymentDeletingState,
}

// flinkDeletableOrCancelableDeploymentStates returns the list of Flink Application Deployment states that are
// considered deletable or cancelable.
var flinkDeletableOrCancelableDeploymentStates = []string{
	flinkDeploymentCreatedState,
	flinkDeploymentRunningState,
	flinkRestartingState,
	flinkDeploymentFailedState,
	flinkDeploymentCanceledState,
	flinkDeploymentFinishedState,
	flinkDeploymentSuspendedState,
}

// flinkApplicationDeploymentAllowedStateTransformations is the map of allowed state transformations for the Flink
// Application Deployment resource.
//
// The keys are the source states, and the values are the list of target states that are allowed to be transformed to.
// If the target state is not in the list, the state transformation is not allowed.
// If the list of target states is empty, the state transformation is not allowed.
//
// See https://nightlies.apache.org/flink/flink-docs-master/docs/internals/job_scheduling/ for more information about
// the state transitions.
var flinkApplicationDeploymentAllowedStateTransformations = map[string][]string{
	// Initial states.
	flinkDeploymentInitializingState: {
		flinkDeploymentCreatedState,
		flinkDeploymentFailedState,
	},
	flinkDeploymentCreatedState: {
		flinkDeploymentRunningState,
		flinkRestartingState,
		flinkDeploymentFailingState,
		flinkDeploymentCancellingRequestedState,
		flinkDeploymentFailedState,
		flinkDeploymentSuspendedState,
	},

	// Active running states.
	flinkDeploymentRunningState: {
		flinkRestartingState,
		flinkDeploymentFailingState,
		flinkDeploymentCancellingRequestedState,
		flinkDeploymentSavingAndStopRequestedState,
		flinkDeploymentFailedState,
		flinkDeploymentSuspendedState,
	},
	flinkRestartingState: {
		flinkDeploymentRunningState,
		flinkDeploymentFailingState,
		flinkDeploymentCancellingRequestedState,
		flinkDeploymentCancellingState,
		flinkDeploymentFailedState,
		flinkDeploymentCanceledState,
		flinkDeploymentFinishedState,
		flinkDeploymentSuspendedState,
	},
	flinkDeploymentSavingState: {}, // Not implemented yet. No state transformations allowed.

	// Intermediate states.
	flinkDeploymentFailingState: {
		flinkDeploymentRunningState,
		flinkRestartingState,
		flinkDeploymentFailedState,
		flinkDeploymentSuspendedState,
	},
	flinkDeploymentCancellingRequestedState: {
		flinkDeploymentFailingState,
		flinkDeploymentCancellingState,
		flinkDeploymentFailedState,
		flinkDeploymentSuspendedState,
	},
	flinkDeploymentCancellingState: {
		flinkDeploymentFailingState,
		flinkDeploymentFailedState,
		flinkDeploymentCanceledState,
		flinkDeploymentSuspendedState,
	},
	flinkDeploymentSavingAndStopRequestedState: {
		flinkDeploymentFailingState,
		flinkDeploymentCancellingState,
		flinkDeploymentSavingAndStopState,
		flinkDeploymentFailedState,
		flinkDeploymentSuspendedState,
	},
	flinkDeploymentSavingAndStopState: {
		flinkDeploymentFailingState,
		flinkDeploymentFailedState,
		flinkDeploymentFinishedState,
		flinkDeploymentSuspendedState,
	},

	// Terminal states.
	flinkDeploymentFailedState: { // It's possible to get to the failed state from any state.
		flinkDeploymentDeleteRequestedState,
	},
	flinkDeploymentCanceledState: {
		flinkDeploymentDeleteRequestedState,
	},
	flinkDeploymentFinishedState: {
		flinkDeploymentDeleteRequestedState,
	},
	flinkDeploymentSuspendedState: {
		flinkDeploymentRunningState,
		flinkDeploymentFailingState,
		flinkDeploymentCancellingRequestedState,
		flinkDeploymentCancellingState,
		flinkDeploymentFailedState,
		flinkDeploymentCanceledState,
		flinkDeploymentFinishedState,
		flinkDeploymentDeleteRequestedState,
	},
	flinkDeploymentDeleteRequestedState: {
		flinkDeploymentDeletingState,
	},
	flinkDeploymentDeletingState: {}, // No state transformations allowed.
}

// aivenFlinkApplicationDeploymentSchema is the schema for the Flink Application Deployment resource.
var aivenFlinkApplicationDeploymentSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"application_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Application ID",
	},
	// Request fields.
	"parallelism": {
		Type:         schema.TypeInt,
		Optional:     true,
		Description:  "Flink Job parallelism",
		ValidateFunc: validation.IntBetween(1, 128),
		ForceNew:     true,
		Default:      1,
	},
	"restart_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Specifies whether a Flink Job is restarted in case it fails",
		ForceNew:    true,
		Default:     true,
	},
	"starting_savepoint": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Job savepoint",
		ValidateFunc: validation.StringLenBetween(1, 2048),
		ForceNew:     true,
	},
	"version_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "ApplicationVersion ID",
		ForceNew:    true,
	},
	// Computed fields.
	"created_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Application deployment creation time",
	},
	"created_by": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Application deployment creator",
	},
}

// ResourceFlinkApplicationDeployment returns the schema for the Flink Application Deployment resource.
func ResourceFlinkApplicationDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "The Flink Application Deployment resource allows the creation and management of Aiven Flink " +
			"Application Deployments.",
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

	gr, err := waitForStateContext(
		ctx, client,
		project, serviceName, applicationID, r.ID,
		flinkInitialDeploymentStates,
		append(flinkActiveDeploymentStates, flinkInactiveDeploymentStates...),
		d.Timeout(schema.TimeoutCreate),
	)

	// If the deployment is restarting or inactive right after creation, we need to cancel it and return an error.
	// This is because the deployment should not be restarting or inactive right after creation, and it indicates an
	// error in the SQL code or misconfiguration.
	if err == nil {
		d.SetId(schemautil.BuildResourceID(project, serviceName, applicationID, r.ID))

		// We call resourceFlinkApplicationDeploymentDelete directly to avoid duplicating the deletion logic.
		// This is the reason why ID is set above, as it is needed in the resourceFlinkApplicationDeploymentDelete.
		if gr.Status == flinkRestartingState || slices.Contains(flinkInactiveDeploymentStates, gr.Status) {
			diagnostics := resourceFlinkApplicationDeploymentDelete(ctx, d, m)
			if diagnostics.HasError() {
				// This should never happen, but just in case, we return the diagnostics that were returned by the
				// deletion function.
				return diagnostics
			}

			err = errors.New(
				"flink application deployment is restarting or inactive when it was not expected to; check your " +
					"SQL and config for errors, see flink logs for more information",
			)
		}
	}

	if err != nil {
		return diag.Errorf("error waiting for Flink Application Deployment to become running: %s", err)
	}

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

	// This is a fix for the problem when the deployment is in such a state that can neither be deleted nor canceled.
	r, err := waitForStateContext(
		ctx, client,
		project, serviceName, applicationID, deploymentID,
		flinkNonDeletableOrCancelableDeploymentStates,
		flinkDeletableOrCancelableDeploymentStates,
		d.Timeout(schema.TimeoutDelete),
	)
	if err != nil {
		var e aiven.Error
		if errors.As(err, &e) && e.Status == 404 {
			// 404 means that the deployment does not exist, so we can consider it deleted.
			return nil
		}

		return diag.Errorf(
			"error waiting for Flink Application Deployment to become deletable or cancelable: %s", err,
		)
	}

	if slices.Contains(
		flinkApplicationDeploymentAllowedStateTransformations[r.Status], flinkDeploymentCancellingRequestedState,
	) {
		_, err = client.FlinkApplicationDeployments.Cancel(ctx, project, serviceName, applicationID, deploymentID)
		if err != nil {
			return diag.Errorf("error cancelling Flink Application Deployment: %v", err)
		}

		_, err = waitForStateContext(
			ctx, client,
			project, serviceName, applicationID, deploymentID,
			[]string{
				flinkDeploymentCancellingRequestedState,
				flinkDeploymentCancellingState,
			},
			[]string{
				flinkDeploymentCanceledState,
			},
			d.Timeout(schema.TimeoutDelete),
		)
		if err != nil {
			return diag.Errorf("error waiting for Flink Application Deployment to become canceled: %s", err)
		}

		r, err = client.FlinkApplicationDeployments.Get(ctx, project, serviceName, applicationID, deploymentID)
		if err != nil {
			return diag.Errorf("cannot get Flink Application Deployment: %v", err)
		}
	}

	if slices.Contains(
		flinkApplicationDeploymentAllowedStateTransformations[r.Status], flinkDeploymentDeleteRequestedState,
	) {
		_, err = client.FlinkApplicationDeployments.Delete(ctx, project, serviceName, applicationID, deploymentID)
		if err != nil {
			return diag.Errorf("error deleting Flink Application Deployment: %v", err)
		}
	}

	return nil
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

// waitForStateContext waits for the Flink Application Deployment to reach the target state.
func waitForStateContext(
	ctx context.Context,
	client *aiven.Client,
	project, serviceName, applicationID, deploymentID string,
	pendingStates, targetStates []string,
	timeout time.Duration,
) (*aiven.GetFlinkApplicationDeploymentResponse, error) {
	conf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (any, string, error) {
			r, err := client.FlinkApplicationDeployments.Get(ctx, project, serviceName, applicationID, deploymentID)
			if err != nil {
				return nil, "", err
			}

			return r, r.Status, nil
		},
		Delay:      1 * time.Second,
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}

	r, err := conf.WaitForStateContext(ctx)
	if r == nil {
		return nil, err
	}

	return r.(*aiven.GetFlinkApplicationDeploymentResponse), err
}
