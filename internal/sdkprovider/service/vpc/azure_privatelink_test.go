package vpc_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAzurePrivatelink_basic(t *testing.T) {
	if os.Getenv("AIVEN_AZURE_PRIVATELINK_VPCID") == "" ||
		os.Getenv("AIVEN_AZURE_PRIVATELINK_SUB_ID") == "" {
		t.Skip("AIVEN_AZURE_PRIVATELINK_VPCID and AIVEN_AZURE_PRIVATELINK_SUB_ID env variables are required to run this test")
	}

	resourceName := "aiven_azure_privatelink.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAzurePrivatelinkResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzurePrivatelinkResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAzurePrivatelinkAttributes("data.aiven_azure_privatelink.pr"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_service_id"),
					resource.TestCheckResourceAttrSet(resourceName, "azure_service_alias"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
				),
			},
		},
	})
}

func testAccCheckAivenAzurePrivatelinkResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each Azure privatelink is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_azure_privatelink" {
			continue
		}

		project, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		pv, err := c.AzurePrivatelink.Get(ctx, project, serviceName)
		var e aiven.Error
		if common.IsCritical(err) && errors.As(err, &e) && e.Status != 500 {
			return fmt.Errorf("error getting a Azure Privatelink: %w", err)
		}

		if pv != nil {
			return fmt.Errorf("azure privatelink (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAzurePrivatelinkResource(name string) string {
	principal := os.Getenv("AIVEN_AZURE_PRIVATELINK_SUB_ID")
	vpcID := os.Getenv("AIVEN_AZURE_PRIVATELINK_VPCID")

	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "azure-westeurope"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  project_vpc_id          = "%s"

  pg_user_config {
    privatelink_access {
      pg        = true
      pgbouncer = true
    }
  }
}

resource "aiven_azure_privatelink" "foo" {
  project               = data.aiven_project.foo.project
  service_name          = aiven_pg.bar.service_name
  user_subscription_ids = ["%s"]
}

data "aiven_azure_privatelink" "pr" {
  project      = data.aiven_project.foo.project
  service_name = aiven_pg.bar.service_name

  depends_on = [aiven_azure_privatelink.foo]
}`, acc.ProjectName(), name, vpcID, principal)
}

func testAccCheckAivenAzurePrivatelinkAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["azure_service_id"] == "" {
			return fmt.Errorf("expected to get an azure_service_id from Aiven")
		}

		if a["azure_service_alias"] == "" {
			return fmt.Errorf("expected to get an azure_service_alias from Aiven")
		}

		if a["state"] == "" {
			return fmt.Errorf("expected to get a state from Aiven")
		}

		return nil
	}
}
