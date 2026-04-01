package securitypluginconfig_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAivenOpenSearchSecurityPluginConfig_basic
// Once enabled, the security plugin cannot be disabled.
// Therefore, each test should create a new service.
func TestAccAivenOpenSearchSecurityPluginConfig_basic(t *testing.T) {
	projectName := acc.ProjectName()
	resourceName := "aiven_opensearch_security_plugin_config.foo"
	datasourceName := "data.aiven_opensearch_security_plugin_config.foo"
	testPassword := "acc-custom-ThisIsATest123^=^"

	t.Run("basic", func(t *testing.T) {
		serviceName := acc.RandName("os")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
			},
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecurityPluginConfigWithDatasource(projectName, serviceName, testPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "admin_password", testPassword),
						resource.TestCheckResourceAttr(resourceName, "available", "true"),
						resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(resourceName, "admin_enabled", "true"),
						resource.TestCheckResourceAttr(datasourceName, "project", projectName),
						resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(datasourceName, "available", "true"),
						resource.TestCheckResourceAttr(datasourceName, "enabled", "true"),
						resource.TestCheckResourceAttr(datasourceName, "admin_enabled", "true"),
						// The password can't be read be read back, so it's not included in the datasource.
						resource.TestCheckNoResourceAttr(datasourceName, "admin_password"),
					),
				},
				{
					Config: testAccSecurityPluginConfigWithDatasource(projectName, serviceName, testPassword+"new"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "admin_password", testPassword+"new"),
						resource.TestCheckNoResourceAttr(datasourceName, "admin_password"),
					),
				},
			},
		})
	})

	t.Run("backward compatibility", func(t *testing.T) {
		serviceName := acc.RandName("os-compat")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck: func() {
				acc.TestAccPreCheck(t)
			},
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig: testAccSecurityPluginConfigWithDatasource(projectName, serviceName, testPassword),
				Checks: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "admin_password", testPassword),
					resource.TestCheckResourceAttr(resourceName, "available", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_enabled", "true"),
					resource.TestCheckResourceAttr(datasourceName, "project", projectName),
					resource.TestCheckResourceAttr(datasourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(datasourceName, "available", "true"),
					resource.TestCheckResourceAttr(datasourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(datasourceName, "admin_enabled", "true"),
				),
				OldProviderVersion: "4.53.0",
			}),
		})
	})
}

func testAccSecurityPluginConfigWithDatasource(project, serviceName, password string) string {
	return fmt.Sprintf(`
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
}

data "aiven_opensearch_security_plugin_config" "foo" {
  project      = aiven_opensearch_security_plugin_config.foo.project
  service_name = aiven_opensearch_security_plugin_config.foo.service_name
}`, project, serviceName, password)
}
