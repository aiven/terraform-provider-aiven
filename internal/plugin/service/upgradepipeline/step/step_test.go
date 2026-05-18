package step_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenUpgradeStep(t *testing.T) {
	acc.SkipIfNotBeta(t)

	const resourceName = "aiven_upgrade_step.foo"

	projectName := acc.ProjectName()
	organizationName := acc.OrganizationName()
	sourceServiceName := acc.RandName("upgrade-step-src-pg")
	destinationServiceName := acc.RandName("upgrade-step-dst-pg")

	sourceServiceIsReady := acc.CreateTestService(
		t,
		projectName,
		sourceServiceName,
		acc.WithServiceType("pg"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)
	destinationServiceIsReady := acc.CreateTestService(
		t,
		projectName,
		destinationServiceName,
		acc.WithServiceType("pg"),
		acc.WithPlan("startup-4"),
		acc.WithCloud("google-europe-west1"),
	)

	baseConfig := testAccUpgradeStepConfig(
		organizationName,
		projectName,
		sourceServiceName,
		destinationServiceName,
		"",
	)
	updatedConfig := testAccUpgradeStepConfig(
		organizationName,
		projectName,
		sourceServiceName,
		destinationServiceName,
		"auto_validation_delay_days = 3",
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenUpgradeStepResourceDestroy,
		Steps: []resource.TestStep{
			{
				// A step connects two existing services and exposes the API-selected defaults
				PreConfig: func() {
					require.NoError(t, <-sourceServiceIsReady)
					require.NoError(t, <-destinationServiceIsReady)
				},
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "step_id"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_id", "data.aiven_organization.org", "id"),
					resource.TestCheckResourceAttr(resourceName, "source_project_name", projectName),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", sourceServiceName),
					resource.TestCheckResourceAttr(resourceName, "destination_project_name", projectName),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", destinationServiceName),
					resource.TestCheckResourceAttr(resourceName, "auto_validation_delay_days", "7"),
				),
			},
			{
				// A freshly created step is stable and doesn't plan follow-up changes
				Config: baseConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				// The automatic validation window can be tuned without replacing the step
				Config: updatedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						acc.ExpectOnlyAttributesChanged(resourceName, "auto_validation_delay_days"),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "step_id"),
					resource.TestCheckResourceAttr(resourceName, "auto_validation_delay_days", "3"),
				),
			},
			{
				// An existing upgrade step can be adopted into Terraform state
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUpgradeStepConfig(
	organizationName string,
	projectName string,
	sourceServiceName string,
	destinationServiceName string,
	autoValidationDelay string,
) string {
	return fmt.Sprintf(`
data "aiven_organization" "org" {
  name = %[1]q
}

resource "aiven_upgrade_step" "foo" {
  organization_id          = data.aiven_organization.org.id
  source_project_name      = %[2]q
  source_service_name      = %[3]q
  destination_project_name = %[2]q
  destination_service_name = %[4]q
  %[5]s
}`, organizationName, projectName, sourceServiceName, destinationServiceName, autoValidationDelay)
}

func testAccCheckAivenUpgradeStepResourceDestroy(s *terraform.State) error {
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_upgrade_step" {
			continue
		}

		organizationID, stepID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = client.UpgradePipelineStepGet(ctx, organizationID, stepID)
		if avngen.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("upgrade step %q still exists", rs.Primary.ID)
	}

	return nil
}
