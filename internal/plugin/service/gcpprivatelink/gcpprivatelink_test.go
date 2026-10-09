package gcpprivatelink_test

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

func TestAccAivenGCPPrivatelink_basic(t *testing.T) {
	skipGCPPrivatelinkTests(t)

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
					resource.TestCheckResourceAttr(resourceName, "state", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "google_service_attachment"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAivenGCPPrivatelink_backwardCompat(t *testing.T) {
	skipGCPPrivatelinkTests(t)

	resourceName := "aiven_gcp_privatelink.foo"
	dataSourceName := "data.aiven_gcp_privatelink.pr"
	projectName := acc.ProjectName()
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	config := testAccGCPPrivatelinkResource(rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckAivenGCPPrivatelinkResourceDestroy,
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
	})
}

func skipGCPPrivatelinkTests(t *testing.T) {
	t.Helper()

	if _, ok := os.LookupEnv("GCP_PRIVATE_LINK_TEST"); !ok {
		t.Skip("environment variable GCP_PRIVATE_LINK_TEST not set")
	}
}

func testAccCheckAivenGCPPrivatelinkResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_gcp_privatelink" {
			continue
		}

		project, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		privatelink, err := c.GCPPrivatelink.Get(ctx, project, serviceName)
		var apiErr aiven.Error
		if common.IsCritical(err) && errors.As(err, &apiErr) && apiErr.Status != 500 {
			return fmt.Errorf("error getting a GCP Privatelink: %w", err)
		}

		if privatelink != nil {
			return fmt.Errorf("GCP Privatelink (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccGCPPrivatelinkResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = %q
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
  project      = aiven_gcp_privatelink.foo.project
  service_name = aiven_gcp_privatelink.foo.service_name
}
`, acc.ProjectName(), name)
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
