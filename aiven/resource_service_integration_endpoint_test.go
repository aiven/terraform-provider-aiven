package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

func TestAccAivenServiceIntegrationEndpoint_basic(t *testing.T) {
	t.Parallel()
	resourceName := "aiven_service_integration_endpoint.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationEndpointResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceEndpointIntegrationAttributes("data.aiven_service_integration_endpoint.endpoint"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", fmt.Sprintf("test-acc-ie-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "datadog"),
				),
			},
		},
	})
}

func testAccServiceIntegrationEndpointResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}
		
		resource "aiven_service" "bar-pg" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-pg-%s"
			service_type = "pg"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				pg_version = 11

				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}
		}
		
		resource "aiven_service_integration_endpoint" "bar" {
			project = aiven_project.foo.project
			endpoint_name = "test-acc-ie-%s"
			endpoint_type = "datadog"
			datadog_user_config {
				datadog_api_key = "Jwx4dl20zOfyYsJGvuv2fiV2VZzCgsuK"
			}
		}

		resource "aiven_service_integration" "bar" {
			project = aiven_project.foo.project
			integration_type = "datadog"
			source_service_name = aiven_service.bar-pg.service_name
			destination_endpoint_id = aiven_service_integration_endpoint.bar.id
		}

		data "aiven_service_integration_endpoint" "endpoint" {
			project = aiven_service_integration_endpoint.bar.project
			endpoint_name = aiven_service_integration_endpoint.bar.endpoint_name
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name, name)
}

func testAccCheckAivenServiceIntegraitonEndpointResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_service_integration_endpoint is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_integration_endpoint" {
			continue
		}

		projectName, endpointId := splitResourceID2(rs.Primary.ID)
		i, err := c.ServiceIntegrationEndpoints.Get(projectName, endpointId)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if i != nil {
			return fmt.Errorf("service integration endpoint(%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAivenServiceEndpointIntegrationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["endpoint_name"] == "" {
			return fmt.Errorf("expected to get a endpoint_name from Aiven")
		}

		if a["endpoint_type"] != "datadog" {
			return fmt.Errorf("expected to get a correct endpoint_type from Aiven")
		}

		return nil
	}
}
