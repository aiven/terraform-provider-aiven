// Package opensearch_test implements tests for the Aiven OpenSearch service.
package opensearch_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// openSearchSecurityPluginTestPassword is the password used for the OpenSearch Security Plugin Config tests.
const openSearchSecurityPluginTestPassword = "ThisIsATest123^=^"

// TestAccAivenOpenSearchSecurityPluginConfig_basic tests the basic functionality of the OpenSearch Security Plugin
// Config resource.
func TestAccAivenOpenSearchSecurityPluginConfig_basic(t *testing.T) {
	resourceName := "aiven_opensearch_security_plugin_config.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_opensearch" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-os-sec-plugin-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_opensearch_user" "foo" {
  service_name = aiven_opensearch.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%[2]s"
}

resource "aiven_opensearch_security_plugin_config" "foo" {
  project        = data.aiven_project.foo.project
  service_name   = aiven_opensearch.bar.service_name
  admin_password = "%s"

  depends_on = [aiven_opensearch.bar, aiven_opensearch_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), rName, openSearchSecurityPluginTestPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(
						resourceName, "service_name", fmt.Sprintf("test-acc-sr-os-sec-plugin-%s", rName),
					),
					resource.TestCheckResourceAttr(
						resourceName, "admin_password", openSearchSecurityPluginTestPassword,
					),
					resource.TestCheckResourceAttr(resourceName, "available", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_enabled", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_opensearch" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-os-sec-plugin-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_opensearch_security_plugin_config" "foo" {
  project        = data.aiven_project.foo.project
  service_name   = aiven_opensearch.bar.service_name
  admin_password = "%s"

  depends_on = [aiven_opensearch.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), rName, openSearchSecurityPluginTestPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(
						resourceName, "service_name", fmt.Sprintf("test-acc-sr-os-sec-plugin-%s", rName),
					),
					resource.TestCheckResourceAttr(
						resourceName, "admin_password", openSearchSecurityPluginTestPassword,
					),
					resource.TestCheckResourceAttr(resourceName, "available", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "admin_enabled", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_opensearch" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-os-sec-plugin-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_opensearch_user" "foo" {
  service_name = aiven_opensearch.bar.service_name
  project      = data.aiven_project.foo.project
  username     = "user-%[2]s"
}

resource "aiven_opensearch_security_plugin_config" "foo" {
  project        = data.aiven_project.foo.project
  service_name   = aiven_opensearch.bar.service_name
  admin_password = "%s"

  depends_on = [aiven_opensearch.bar, aiven_opensearch_user.foo]
}`, os.Getenv("AIVEN_PROJECT_NAME"), rName, openSearchSecurityPluginTestPassword),
				ExpectError: regexp.MustCompile("when the OpenSearch Security Plugin is enabled"),
			},
		},
	})
}
