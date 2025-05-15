package connectionpool

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_connection_pool", &resource.Sweeper{
		Name: "aiven_connection_pool",
		F:    sweepConnectionPoll(ctx),
	})
}

func sweepConnectionPoll(ctx context.Context) func(string) error {
	return func(_ string) error {
		client, err := sweep.SharedClient()
		if err != nil {
			return err
		}

		projectName := os.Getenv("AIVEN_PROJECT_NAME")

		services, err := client.Services.List(ctx, projectName)
		if common.IsCritical(err) {
			return fmt.Errorf("error retrieving a list of services for a project `%s`: %w", projectName, err)
		}

		for _, s := range services {
			switch s.Type {
			case schemautil.ServiceTypeAlloyDBOmni, schemautil.ServiceTypePG:
			default:
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
