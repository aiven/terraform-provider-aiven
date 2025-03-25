package m3db_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAiven_m3aggregator(t *testing.T) {
	resourceName := "aiven_m3aggregator.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccM3AggregatorResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_m3aggregator.common"),
					testAccCheckAivenServiceM3AggregatorAttributes("data.aiven_m3aggregator.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-m3a-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "m3aggregator"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func testAccM3AggregatorResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_m3db" "foo" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-8"
  service_name = "test-acc-m3d-%s"

  m3db_user_config {
    namespaces {
      name = "%s"
      type = "unaggregated"
    }
  }
}

resource "aiven_m3aggregator" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-8"
  service_name            = "test-acc-m3a-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_service_integration" "int-m3db-aggr" {
  project                  = data.aiven_project.foo.project
  integration_type         = "m3aggregator"
  source_service_name      = aiven_m3db.foo.service_name
  destination_service_name = aiven_m3aggregator.bar.service_name
}

data "aiven_m3aggregator" "common" {
  service_name = aiven_m3aggregator.bar.service_name
  project      = aiven_m3aggregator.bar.project

  depends_on = [aiven_m3aggregator.bar]
}`, acc.ProjectName(), name, name, name)
}

func testAccCheckAivenServiceM3AggregatorAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["m3aggregator.0.uris.#"] == "" {
			return fmt.Errorf("expected to get correct uris from Aiven")
		}

		if a["m3aggregator.0.aggregator_http_uri"] == "" {
			return fmt.Errorf("expected to get correct aggregator_http_uri from Aiven")
		}

		return nil
	}
}
