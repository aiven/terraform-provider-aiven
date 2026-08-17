package jardeployment

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Create = createViewRestartEnabled
	ResourceOptions.Delete = deleteView
	ResourceOptions.Read = readViewRestartEnabled
}

// createViewRestartEnabled sends the API default when the configuration omits restart_enabled.
// The schema carries no default (see dropDefault in the definition): SDKv2 stored false for an
// omitted optional bool, and a generated true would replace those deployments.
// todo: drop in v5.0.0 along with dropDefault, and let the schema hold the default.
func createViewRestartEnabled(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if _, ok := d.GetOk("restart_enabled"); !ok {
		if err := d.Set("restart_enabled", true); err != nil {
			return err
		}
	}

	return createView(ctx, client, d)
}

// readViewRestartEnabled fills restart_enabled, which the API accepts on create but never returns.
// An imported deployment would otherwise hold a null. Existing state is left as it is:
// the stored value is the one the deployment was created with.
//
// An import can only assume the API default, so importing a deployment created with
// restart_enabled = false and declaring that value plans a replacement. Storing a null instead
// would plan one for either value, since the attribute forces replacement when it changes.
func readViewRestartEnabled(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := readView(ctx, client, d)
	if err != nil {
		return err
	}

	// A value stored by a previous apply is the one the deployment was created with.
	if _, ok := d.GetOk("restart_enabled"); ok {
		return nil
	}

	return d.Set("restart_enabled", true)
}

// deleteView handles the complex state machine for Flink Jar Application Deployment deletion.
// The deployment must be canceled before it can be deleted.
// Retries until the deployment is gone or the context times out.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationID := d.Get("application_id").(string)
	deploymentID := d.Get("deployment_id").(string)

	// Flink Jar Application Deployment has a quite complicated state machine
	// https://api.aiven.io/doc/#tag/Service:_Flink/operation/ServiceFlinkGetJarApplicationDeployment
	// Retries until succeeds or exceeds the timeout
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
			return fmt.Errorf("can't delete Flink Jar Application Deployment: %w", ctx.Err())
		case <-time.After(time.Second):
			continue
		}
	}
}
