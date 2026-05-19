package step

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_upgrade_step", &resource.Sweeper{
		Name: "aiven_upgrade_step",
		F: func(_ string) error {
			orgName := os.Getenv("AIVEN_ORGANIZATION_NAME")
			projectName := os.Getenv("AIVEN_PROJECT_NAME")

			client, err := sweep.SharedGenClient()
			if err != nil {
				return err
			}

			organizations, err := client.AccountList(ctx)
			if err != nil {
				return fmt.Errorf("error retrieving a list of organizations: %w", err)
			}

			for _, organization := range organizations {
				if organization.AccountName != orgName {
					continue
				}

				steps, err := client.UpgradePipelineStepList(ctx, organization.OrganizationId)
				if err != nil {
					return fmt.Errorf("error retrieving upgrade steps for organization %s: %w", organization.OrganizationId, err)
				}

				if steps == nil {
					continue
				}

				for _, step := range steps.Steps {
					// This assumes acceptance tests create upgrade steps within AIVEN_PROJECT_NAME.
					// Cross-project step coverage needs a broader project ownership check here.
					if step.SourceProjectName != projectName && step.DestinationProjectName != projectName {
						continue
					}

					if !strings.HasPrefix(step.SourceServiceName, sweep.DefaultPrefix) &&
						!strings.HasPrefix(step.DestinationServiceName, sweep.DefaultPrefix) {
						continue
					}

					if err = client.UpgradePipelineStepDelete(ctx, organization.OrganizationId, step.StepId); common.IsCritical(err) {
						return fmt.Errorf("error deleting upgrade step %s: %w", step.StepId, err)
					}
				}
			}

			return nil
		},
	})
}
