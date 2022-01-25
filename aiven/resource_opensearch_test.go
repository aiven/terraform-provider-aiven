// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
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
	t.Parallel()

	resourceName := "aiven_opensearch.bar-os"
	projectName := os.Getenv("AIVEN_PROJECT_NAME")

	t.Run("basic resource", func(tt *testing.T) {
		serviceName := fmt.Sprintf("test-acc-os-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

		manifest := fmt.Sprintf(`
			data "aiven_project" "foo-es" {
			  project = "%s"
			}
			
			resource "aiven_opensearch" "bar-os" {
			  project                 = data.aiven_project.foo-es.project
			  cloud_name              = "google-europe-west1"
			  plan                    = "startup-4"
			  service_name            = "%s"
			  maintenance_window_dow  = "monday"
			  maintenance_window_time = "10:00:00"
			
			  opensearch_user_config {
			    opensearch_dashboards {
			      enabled = true
			    }
			
			    public_access {
			      opensearch            = true
			      opensearch_dashboards = true
			    }
			
			    index_patterns {
			      pattern           = "logs_*_foo_*"
			      max_index_count   = 3
			      sorting_algorithm = "creation_date"
			    }
			
			    index_patterns {
			      pattern           = "logs_*_bar_*"
			      max_index_count   = 15
			      sorting_algorithm = "creation_date"
			    }
			  }
			}
			
			data "aiven_opensearch" "service-os" {
			  service_name = aiven_opensearch.bar-os.service_name
			  project      = aiven_opensearch.bar-os.project
			
			  depends_on = [aiven_opensearch.bar-os]
			}`,
			projectName, serviceName)

		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: manifest,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServiceCommonAttributes("data.aiven_opensearch.service-os"),
						testAccCheckAivenServiceOSAttributes("data.aiven_opensearch.service-os"),
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "project", projectName),
						resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
						resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
						resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
						resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					),
				},
			},
		})
	})

	t.Run("migrate elasticsearch resource to opensearch resource", func(tt *testing.T) {
		serviceName := fmt.Sprintf("test-acc-es-to-os-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

		esManifestWithVersion := func(version string) string {
			return fmt.Sprintf(`
				resource "aiven_elasticsearch" "import-me" {
				  project                 = "%s"
				  cloud_name              = "google-europe-west1"
				  plan                    = "startup-4"
				  service_name            = "%s"
				  maintenance_window_dow  = "monday"
				  maintenance_window_time = "10:00:00"
				
				  elasticsearch_user_config {
				    elasticsearch_version = parseint("%s", 10)
				
				    index_patterns {
				      pattern           = "logs_*_foo_*"
				      max_index_count   = 3
				      sorting_algorithm = "creation_date"
				    }
				
				    index_patterns {
				      pattern           = "logs_*_bar_*"
				      max_index_count   = 15
				      sorting_algorithm = "creation_date"
				    }
				  }
				}`,
				projectName, serviceName, version)

		}

		importManifest := fmt.Sprintf(`
			resource "aiven_opensearch" "bar-os" {
			  project                 = "%s"
			  cloud_name              = "google-europe-west1"
			  plan                    = "startup-4"
			  service_name            = "%s"
			  maintenance_window_dow  = "monday"
			  maintenance_window_time = "10:00:00"
			
			  opensearch_user_config {
			    opensearch_version = 1
			
			    index_patterns {
			      pattern           = "logs_*_foo_*"
			      max_index_count   = 3
			      sorting_algorithm = "creation_date"
			    }
			
			    index_patterns {
			      pattern           = "logs_*_bar_*"
			      max_index_count   = 15
			      sorting_algorithm = "creation_date"
			    }
			  }
			}`,
			projectName, serviceName)

		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,

			Steps: []resource.TestStep{
				{Config: esManifestWithVersion("7")},
				{Config: esManifestWithVersion("1")},
				{
					Config:        importManifest,
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: fmt.Sprintf("%s/%s", projectName, serviceName),
					ImportStateCheck: func(s []*terraform.InstanceState) error {
						if len(s) != 1 {
							return fmt.Errorf("expected only one instance to be imported, state: %#v", s)
						}
						rs := s[0]
						attributes := rs.Attributes
						if !strings.EqualFold(attributes["service_type"], "opensearch") {
							return fmt.Errorf("expected service_type opensearch after import, got :%s", attributes["service_type"])
						}
						if !strings.EqualFold(attributes["service_name"], serviceName) {
							return fmt.Errorf("expected service_name to match '%s', got: '%s'", serviceName, attributes["service_name"])
						}
						if !strings.EqualFold(attributes["project"], projectName) {
							return fmt.Errorf("expected project to match '%s', got: '%s'", serviceName, attributes["project_name"])
						}
						return nil
					},
				},
			},
		})
	})
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
