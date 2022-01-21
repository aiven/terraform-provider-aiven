// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAiven_cassandra(t *testing.T) {
	resourceName := "aiven_cassandra.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCassandraResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_cassandra.service"),
					testAccCheckAivenServiceCassandraAttributes("data.aiven_cassandra.service"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "cassandra"),
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

func testAccCassandraResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_cassandra" "bar" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  cassandra_user_config {
		    migrate_sstableloader = true
		
		    public_access {
		      prometheus = true
		    }
		  }
		}
		
		data "aiven_cassandra" "service" {
		  service_name = aiven_cassandra.bar.service_name
		  project      = data.aiven_project.foo.project
		
		  depends_on = [aiven_cassandra.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name)
}
