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

func TestAccAiven_elasticsearch(t *testing.T) {
	resourceName := "aiven_elasticsearch.bar-es"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceCommonAttributes("data.aiven_elasticsearch.service-es"),
					testAccCheckAivenServiceESAttributes("data.aiven_elasticsearch.service-es"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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

func testAccElasticsearchResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo-es" {
		  project = "%s"
		}
		
		resource "aiven_elasticsearch" "bar-es" {
		  project                 = data.aiven_project.foo-es.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  elasticsearch_user_config {
		    kibana {
		      enabled                       = true
		      elasticsearch_request_timeout = 30000
		    }
		
		    public_access {
		      elasticsearch = true
		      kibana        = true
		    }
		  }
		}
		
		data "aiven_elasticsearch" "service-es" {
		  service_name = aiven_elasticsearch.bar-es.service_name
		  project      = data.aiven_project.foo-es.project
		
		  depends_on = [aiven_elasticsearch.bar-es]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name)
}
