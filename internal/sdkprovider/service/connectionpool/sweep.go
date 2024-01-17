package connectionpool

import (
	"context"
	"fmt"
	"os"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	if os.Getenv("TF_SWEEP") == "" {
		return
	}

	ctx := context.Background()

	client, err := sweep.SharedClient()
	if err != nil {
		panic(fmt.Sprintf("error getting client: %s", err))
	}

	sweep.AddTestSweepers("aiven_connection_pool", &resource.Sweeper{
		Name: "aiven_connection_pool",
		F:    sweepConnectionPoll(ctx, client),
	})

}

func sweepConnectionPoll(ctx context.Context, client *aiven.Client) func(string) error {
	return func(id string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")

		services, err := client.Services.List(ctx, projectName)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of services for a project `%s`: %w", projectName, err)
		}

		for _, s := range services {
			if s.Type != schemautil.ServiceTypePG {
				continue
			}

			l, err := client.ConnectionPools.List(ctx, projectName, s.Name)
			if common.IsCritical(err) {
				return err
			}

			for _, pool := range l {
				err = client.ConnectionPools.Delete(ctx, projectName, s.Name, pool.PoolName)
				if common.IsCritical(err) {
					return err
				}
			}
		}
		return nil
	}
}
