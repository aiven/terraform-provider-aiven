// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Opensearch service tests
func TestAccAivenService_os(t *testing.T) {
	resourceName := "aiven_opensearch.bar-os"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpensearchServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_opensearch.service-os"),
					testAccCheckAivenServiceOSAttributes("data.aiven_opensearch.service-os"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-os-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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

func testAccOpensearchServiceResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo-es" {
			project = "%s"
		}
		
		resource "aiven_opensearch" "bar-os" {
			project = data.aiven_project.foo-es.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-os-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			opensearch_user_config {
				opensearch_dashboards {
					enabled = true
				}
	
				public_access {
					opensearch = true
					opensearch_dashboards = true
				}

				index_patterns {
					pattern = "logs_*_foo_*"
					max_index_count = 3
					sorting_algorithm = "creation_date"
				}

				index_patterns {
					pattern = "logs_*_bar_*"
					max_index_count = 15
					sorting_algorithm = "creation_date"
				}
			}
		}
		
		data "aiven_opensearch" "service-os" {
			service_name = aiven_opensearch.bar-os.service_name
			project = aiven_opensearch.bar-os.project

			depends_on = [aiven_opensearch.bar-os]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceOSAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "opensearch") {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["opensearch_dashboards_uri"] != "" {
			return fmt.Errorf("expected opensearch_dashboards_uri to not be empty")
		}

		if a["opensearch_user_config.0.ip_filter.0"] != "0.0.0.0/0" {
			return fmt.Errorf("expected to get a correct ip_filter from Aiven")
		}

		if a["opensearch_user_config.0.public_access.0.opensearch"] != "true" {
			return fmt.Errorf("expected to get opensearch.public_access enabled from Aiven")
		}

		if a["opensearch_user_config.0.public_access.0.prometheus"] != "" {
			return fmt.Errorf("expected to get a correct public_access prometheus from Aiven")
		}

		return nil
	}
}
