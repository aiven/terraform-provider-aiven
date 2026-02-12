package deployment

import (
	"context"
	"fmt"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func init() {
	ResourceOptions.Delete = deleteView
}

// planModifier extracts deployment_id from the composite ID if it's not set,
// which happens during migration from the SDK provider that didn't store deployment_id as a separate attribute.
func planModifier(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	if d.Get("deployment_id").(string) == "" && d.ID() != "" {
		parts := strings.SplitN(d.ID(), "/", 4)
		if len(parts) == 4 {
			return d.Set("deployment_id", parts[3])
		}
	}
	return nil
}

// deleteView handles the complex state machine for Flink Application Deployment deletion.
// The deployment must be canceled before it can be deleted.
// Retries until the deployment is gone or the context times out.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationID := d.Get("application_id").(string)
	deploymentID := d.Get("deployment_id").(string)

	// Flink Application Deployment has a quite complicated state machine
	// https://api.aiven.io/doc/#tag/Service:_Flink/operation/ServiceFlinkDeleteApplicationDeployment
	// Retries until succeeds or exceeds the timeout
	for {
		_, err := client.ServiceFlinkGetApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		if avngen.IsNotFound(err) {
			return nil
		}

		// Must be canceled before deleted
		_, err = client.ServiceFlinkCancelApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		if err != nil {
			// Completely ignores all errors, until it gets 404 on GET request
			_, _ = client.ServiceFlinkDeleteApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("can't delete Flink Application Deployment: %w", ctx.Err())
		case <-time.After(time.Second):
			continue
		}
	}
}
