package vpc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenGCPPrivatelink_basic(t *testing.T) {
	t.Skip("Skipping GCP Privatelink tests temporarily due to the ACL issue") // TODO: enable it when the issue is fixed

	resourceName := "aiven_gcp_privatelink.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenGCPPrivatelinkResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPPrivatelinkResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenGCPPrivatelinkAttributes("data.aiven_gcp_privatelink.pr"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "google_service_attachment"),
				),
			},
		},
	})
}

func testAccCheckAivenGCPPrivatelinkResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each GCP privatelink is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_gcp_privatelink" {
			continue
		}

		project, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		pv, err := c.GCPPrivatelink.Get(ctx, project, serviceName)
		var e aiven.Error
		if common.IsCritical(err) && errors.As(err, &e) && e.Status != 500 {
			return fmt.Errorf("error getting a GCP Privatelink: %w", err)
		}

		if pv != nil {
			return fmt.Errorf("gcp privatelink (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccGCPPrivatelinkResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aiven_kafka" "bar" {
  project        = data.aiven_project.foo.project
  cloud_name     = "google-europe-west1"
  plan           = "business-4"
  service_name   = "test-acc-sr-%s"
  project_vpc_id = aiven_project_vpc.aiven_vpc.id

  kafka_user_config {
    privatelink_access {
      kafka         = true
      kafka_connect = true
      kafka_rest    = true
    }
  }
}

resource "aiven_gcp_privatelink" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
}

data "aiven_gcp_privatelink" "pr" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name

  depends_on = [aiven_gcp_privatelink.foo]
}`, acc.ProjectName(), name)
}

func testAccCheckAivenGCPPrivatelinkAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["state"] == "" {
			return fmt.Errorf("expected to get a state from Aiven")
		}

		if a["google_service_attachment"] == "" {
			return fmt.Errorf("expected to get a google_service_attachment from Aiven")
		}

		return nil
	}
}
