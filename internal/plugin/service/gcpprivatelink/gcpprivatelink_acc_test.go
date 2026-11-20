package gcpprivatelink_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenGCPPrivatelink_basic(t *testing.T) {
	acc.RequireEnvVars(t, "GCP_PRIVATE_LINK_TEST")

	resourceName := "aiven_gcp_privatelink.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	config := testAccGCPPrivatelinkResource(rName)
	var identityImport acc.IdentityImportVersionCheck
	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckGCPPrivatelinkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenGCPPrivatelinkAttributes("data.aiven_gcp_privatelink.pr"),
					resource.TestCheckResourceAttr(resourceName, "state", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "google_service_attachment"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resourceName, ImportState: true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
			},
		},
	})
}

func TestAccAivenGCPPrivatelink_backwardCompat(t *testing.T) {
	acc.RequireEnvVars(t, "GCP_PRIVATE_LINK_TEST")

	resourceName := "aiven_gcp_privatelink.foo"
	dataSourceName := "data.aiven_gcp_privatelink.pr"
	projectName := acc.ProjectName()
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	config := testAccGCPPrivatelinkResource(rName)

	var identityImport acc.IdentityImportVersionCheck
	testCase := resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
		PreCheck:               func() { acc.TestAccPreCheck(t) },
		CheckDestroy:           acc.TestAccCheckGCPPrivatelinkDestroyed,
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.60.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "project", projectName),
				resource.TestCheckResourceAttr(resourceName, "service_name", "test-acc-sr-"+rName),
				resource.TestCheckResourceAttr(resourceName, "state", "active"),
				resource.TestCheckResourceAttrSet(resourceName, "google_service_attachment"),
				resource.TestCheckResourceAttrPair(dataSourceName, "project", resourceName, "project"),
				resource.TestCheckResourceAttrPair(dataSourceName, "service_name", resourceName, "service_name"),
				resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
				resource.TestCheckResourceAttrPair(dataSourceName, "google_service_attachment", resourceName, "google_service_attachment"),
			),
		}),
	}
	testCase.Steps[1].ConfigPlanChecks.PreApply = []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}
	testCase.Steps = append(testCase.Steps, acc.ProviderRollbackSteps(t, testCase.Steps[0], testCase.Steps[1], &identityImport)...)
	resource.Test(t, testCase)
}

func testAccGCPPrivatelinkResource(name string) string {
	return acc.GCPPrivatelinkFixture(name) + `
data "aiven_gcp_privatelink" "pr" {
  project      = aiven_gcp_privatelink.foo.project
  service_name = aiven_gcp_privatelink.foo.service_name
}
`
}

func testAccCheckAivenGCPPrivatelinkAttributes(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributes := s.RootModule().Resources[name].Primary.Attributes

		for _, attribute := range []string{"project", "service_name", "state", "google_service_attachment"} {
			if attributes[attribute] == "" {
				return fmt.Errorf("expected %q to be set", attribute)
			}
		}

		return nil
	}
}
