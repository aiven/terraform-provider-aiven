package connectionpool

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_connection_pool", &resource.Sweeper{
		Name: "aiven_connection_pool",
		F: func(_ string) error {
			client, err := sweep.SharedGenClient()
			if err != nil {
				return err
			}

			projectName := sweep.ProjectName()

			services, err := client.ServiceList(ctx, projectName)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %w", projectName, err)
			}

			for _, s := range services {
				if s.ServiceType != schemautil.ServiceTypePG {
					continue
				}

				for _, pool := range s.ConnectionPools {
					err = client.ServicePGBouncerDelete(ctx, projectName, s.ServiceName, pool.PoolName)
					if common.IsCritical(err) {
						return err
					}
				}
			}

			return nil
		},
	})
}
