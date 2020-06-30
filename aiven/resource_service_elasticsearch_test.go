package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"os"
	"strings"
	"testing"
)

// Elasticsearch service tests
func TestAccAivenService_es(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_service.bar-es"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_service.service-es"),
					testAccCheckAivenServiceESAttributes("data.aiven_service.service-es"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-es-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_type", "elasticsearch"),
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

func testAccElasticsearchServiceResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo-es" {
			project = "test-acc-pr-es-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "bar-es" {
			project = aiven_project.foo-es.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "elasticsearch"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			elasticsearch_user_config {
				elasticsearch_version = 7

				kibana {
					enabled = true
					elasticsearch_request_timeout = 30000
				}

				public_access {
					elasticsearch = true
					kibana = true
				}
			}
		}
		
		data "aiven_service" "service-es" {
			service_name = aiven_service.bar-es.service_name
			project = aiven_project.foo-es.project
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name)
}

func testAccCheckAivenServiceESAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "elasticsearch") {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["elasticsearch.0.kibana_uri"] == "" {
			return fmt.Errorf("expected to get kibana_uri from Aiven")
		}

		if a["elasticsearch_user_config.0.elasticsearch_version"] != "7" {
			return fmt.Errorf("expected to get a correct elasticsearch_version from Aiven")
		}

		if a["elasticsearch_user_config.0.ip_filter.0"] != "0.0.0.0/0" {
			return fmt.Errorf("expected to get a correct ip_filter from Aiven")
		}

		if a["elasticsearch.0.kibana_uri"] == "" {
			return fmt.Errorf("expected to get kibana_uri from Aiven")
		}

		if a["elasticsearch_user_config.0.kibana.0.enabled"] != "true" {
			return fmt.Errorf("expected to get a correct kibana.enabled from Aiven")
		}

		if a["elasticsearch_user_config.0.kibana.0.elasticsearch_request_timeout"] != "30000" {
			return fmt.Errorf("expected to get kibana.elasticsearch_request_timeout from Aiven")
		}

		if a["elasticsearch_user_config.0.public_access.0.elasticsearch"] != "true" {
			return fmt.Errorf("expected to get elasticsearch.public_access enabled from Aiven")
		}

		if a["elasticsearch_user_config.0.public_access.0.kibana"] != "true" {
			return fmt.Errorf("expected to get elasticsearch.public_access enabled for Kibana from Aiven")
		}

		if a["elasticsearch_user_config.0.public_access.0.prometheus"] != "" {
			return fmt.Errorf("expected to get a correct public_access prometheus from Aiven")
		}

		return nil
	}
}
