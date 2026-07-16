package azureprivatelink

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const azurePrivatelinkDeletedState = "deleted"

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return deleteViewInternal(ctx, client, d, common.DefaultStateChangeDelay)
}

// deleteViewInternal deletes an Azure PrivateLink and waits until the API reports it missing.
func deleteViewInternal(ctx context.Context, client avngen.Client, d adapter.ResourceData, retryDelay time.Duration) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	if _, err := client.ServicePrivatelinkAzureDelete(ctx, project, serviceName); err != nil {
		if avngen.IsNotFound(err) {
			return nil
		}
		return err
	}

	for {
		rsp, err := client.ServicePrivatelinkAzureGet(ctx, project, serviceName)
		if err != nil {
			if avngen.IsNotFound(err) {
				return nil
			}
			return err
		}
		if string(rsp.State) == azurePrivatelinkDeletedState {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"azure PrivateLink deletion is pending: service %q is in state %q, waiting for %q: %w",
				serviceName,
				rsp.State,
				azurePrivatelinkDeletedState,
				ctx.Err(),
			)
		case <-time.After(retryDelay):
		}
	}
}
