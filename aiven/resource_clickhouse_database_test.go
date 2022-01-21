// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/

package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAivenClickhouseDatabase_basic(t *testing.T) {
	resourceName := "aiven_clickhouse_database.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenDatabaseResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClickhouseDatabaseResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-db-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func testAccClickhouseDatabaseResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_clickhouse" "bar" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-8"
		  service_name            = "test-acc-sr-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		}
		
		resource "aiven_clickhouse_database" "foo" {
		  project      = aiven_clickhouse.bar.project
		  service_name = aiven_clickhouse.bar.service_name
		  name         = "test-acc-db-%s"
		}
		
		data "aiven_clickhouse_database" "database" {
		  project      = aiven_clickhouse_database.foo.project
		  service_name = aiven_clickhouse_database.foo.service_name
		  name         = aiven_clickhouse_database.foo.name
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}
