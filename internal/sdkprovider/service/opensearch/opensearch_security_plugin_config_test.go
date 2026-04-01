// Package opensearch_test implements tests for the Aiven OpenSearch service.
package opensearch_test

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// openSearchTestPassword is the password used for the OpenSearch tests.
const openSearchTestPassword = "acc-custom-ThisIsATest123^=^"

// TestAccAivenOpenSearchSecurityPluginConfig_basic tests the basic functionality of the OpenSearch Security Plugin
// Config resource.
func TestAccAivenOpenSearchSecurityPluginConfig_basic(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := acc.RandName("opensearch")
	resourceName := "aiven_opensearch_security_plugin_config.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "aiven_opensearch" "bar" {
  project                 = %[1]q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = %[2]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_opensearch_security_plugin_config" "foo" {
  project        = aiven_opensearch.bar.project
  service_name   = aiven_opensearch.bar.service_name
  admin_password = %[3]q
}`, projectName, serviceName, openSearchTestPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "admin_password", openSearchTestPassword),
					resource.TestCheckResourceAttr(resourceName, "available", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckAivenOpenSearchUserResourceDestroy(s *terraform.State) error {
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return fmt.Errorf("error instantiating client: %w", err)
	}

	ctx := context.Background()

	// loop through the resources in state, verifying each aiven_opensearch_user is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_opensearch_user" {
			continue
		}

		projectName, serviceName, username, err := schemautil.SplitResourceID3(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = c.ServiceUserGet(ctx, projectName, serviceName, username)
		if err != nil && !avngen.IsNotFound(err) {
			return fmt.Errorf("error checking if user was destroyed: %w", err)
		}

		if err == nil {
			return fmt.Errorf("opensearch user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
