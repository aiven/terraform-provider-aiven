// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// PG service tests
func TestAccAivenService_pg(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_pg.bar-pg"

	t.Run("basic pg resource", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPGServiceResource(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServiceCommonAttributes("data.aiven_pg.service-pg"),
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service-pg"),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
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
	})

	t.Run("custom timeouts pg resource", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPGServiceCustomTimeoutsResource(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServiceCommonAttributes("data.aiven_pg.service-pg"),
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service-pg"),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
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
	})

	t.Run("read replica pg resource", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config:                    testAccPGReadReplicaServiceResource(rName),
					PreventPostDestroyRefresh: true,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServiceCommonAttributes("data.aiven_pg.service-pg"),
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service-pg"),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
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
	})

	t.Run("termination protection test", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPGTerminationProtectionServiceResource(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServiceTerminationProtection("data.aiven_pg.service-pg"),
						testAccCheckAivenServiceCommonAttributes("data.aiven_pg.service-pg"),
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service-pg"),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
						resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
						resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
						resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
						resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
						resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
					),
					ExpectNonEmptyPlan: true,
				},
			},
		})
	})
}

func testAccPGServiceResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo-pg" {
			project = "%s"
		}
		
		resource "aiven_pg" "bar-pg" {
			project = data.aiven_project.foo-pg.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}
		}
		
		data "aiven_pg" "service-pg" {
			service_name = aiven_pg.bar-pg.service_name
			project = aiven_pg.bar-pg.project

			depends_on = [aiven_pg.bar-pg]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccPGServiceCustomTimeoutsResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo-pg" {
			project = "%s"
		}
		
		resource "aiven_pg" "bar-pg" {
			project = data.aiven_project.foo-pg.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			timeouts {
				create = "25m"
				update = "30m"
			}
			
			pg_user_config {
				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}
		}
		
		data "aiven_pg" "service-pg" {
			service_name = aiven_pg.bar-pg.service_name
			project = aiven_pg.bar-pg.project

			depends_on = [aiven_pg.bar-pg]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccPGTerminationProtectionServiceResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo-pg" {
			project = "%s"
		}
		
		resource "aiven_pg" "bar-pg" {
			project = data.aiven_project.foo-pg.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			termination_protection = true
			
			pg_user_config {
				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}
		}
		
		data "aiven_pg" "service-pg" {
			service_name = aiven_pg.bar-pg.service_name
			project = aiven_pg.bar-pg.project

			depends_on = [aiven_pg.bar-pg]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccPGReadReplicaServiceResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo-pg" {
			project = "%s"
		}
		
		resource "aiven_pg" "bar-pg" {
			project = data.aiven_project.foo-pg.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}
		}

		resource "aiven_pg" "bar-replica" {
			project = data.aiven_project.foo-pg.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-repica-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				backup_hour = 19
				backup_minute = 30
				public_access {
					pg = true
					prometheus = false
				}

				pg {
					idle_in_transaction_session_timeout = 900
				}
			}

			service_integrations {
				integration_type = "read_replica"
				source_service_name = aiven_pg.bar-pg.service_name
			}

			 depends_on = [aiven_pg.bar-pg]
		}

		resource "aiven_service_integration" "pg-readreplica" {
			project = data.aiven_project.foo-pg.project
			integration_type = "read_replica"
			source_service_name = aiven_pg.bar-pg.service_name
			destination_service_name = aiven_pg.bar-replica.service_name

			depends_on = [aiven_pg.bar-replica]
		}
		
		data "aiven_pg" "service-pg" {
			service_name = aiven_pg.bar-pg.service_name
			project = aiven_pg.bar-pg.project

			depends_on = [aiven_pg.bar-pg]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccCheckAivenServiceTerminationProtection(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		projectName, serviceName := splitResourceID2(a["id"])

		c := testAccProvider.Meta().(*aiven.Client)

		service, err := c.Services.Get(projectName, serviceName)
		if err != nil {
			return fmt.Errorf("cannot get service %s err: %s", serviceName, err)
		}

		if service.TerminationProtection == false {
			return fmt.Errorf("expected to get a termination_protection=true from Aiven")
		}

		// try to delete Aiven service with termination_protection enabled
		// should be an error from Aiven API
		err = c.Services.Delete(projectName, serviceName)
		if err == nil {
			return fmt.Errorf("termination_protection enabled should prevent from deletion of a service, deletion went OK")
		}

		// set service termination_protection to false to make Terraform Destroy plan work
		_, err = c.Services.Update(projectName, service.Name, aiven.UpdateServiceRequest{
			Cloud:                 service.CloudName,
			MaintenanceWindow:     &service.MaintenanceWindow,
			Plan:                  service.Plan,
			ProjectVPCID:          service.ProjectVPCID,
			Powered:               true,
			TerminationProtection: false,
			UserConfig:            service.UserConfig,
		})

		if err != nil {
			return fmt.Errorf("unable to update Aiven service to set termination_protection=false err: %s", err)
		}

		return nil
	}
}

func testAccCheckAivenServicePGAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "pg") {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["pg.0.dbname"] != "defaultdb" {
			return fmt.Errorf("expected to get a correct PG dbname from Aiven, got :%s", a["pg.0.dbname"])
		}

		if a["pg_user_config.0.pg.0.idle_in_transaction_session_timeout"] != "900" {
			return fmt.Errorf("expected to get a correct PG idle_in_transaction_session_timeout from Aiven")
		}

		if a["pg_user_config.0.public_access.0.pg"] != "true" {
			return fmt.Errorf("expected to get a correct PG public_access from Aiven")
		}

		if a["pg_user_config.0.public_access.0.pgbouncer"] != "" {
			return fmt.Errorf("expected to get a correct PG public_access from Aiven")
		}

		if a["pg_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct PG public_access from Aiven")
		}

		if a["pg.0.host"] == "" {
			return fmt.Errorf("expected to get a PG host from Aiven")
		}

		if a["pg.0.user"] == "" {
			return fmt.Errorf("expected to get a PG user from Aiven")
		}

		if a["pg.0.password"] == "" {
			return fmt.Errorf("expected to get a PG password from Aiven")
		}

		if a["pg.0.port"] == "" {
			return fmt.Errorf("expected to get a PG port from Aiven")
		}

		if a["pg.0.uri"] == "" {
			return fmt.Errorf("expected to get a PG URI from Aiven")
		}

		if a["pg_user_config.0.service_to_fork_from"] != "" {
			return fmt.Errorf("expected to get a PG service_to_fork_from not set to any value")
		}

		return nil
	}
}
