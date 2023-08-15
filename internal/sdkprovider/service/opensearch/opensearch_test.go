package opensearch_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acctest3 "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// OpenSearch service tests
func TestAccAivenService_os(t *testing.T) {
	resourceName := "aiven_opensearch.bar-os"
	projectName := os.Getenv("AIVEN_PROJECT_NAME")
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

  tag {
    key   = "test"
    value = "val"
  }

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

data "aiven_opensearch" "common-os" {
  service_name = aiven_opensearch.bar-os.service_name
  project      = aiven_opensearch.bar-os.project

  depends_on = [aiven_opensearch.bar-os]
}`, projectName, serviceName)
	manifestDoubleTag := fmt.Sprintf(`
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

  tag {
    key   = "test"
    value = "val"
  }
  tag {
    key   = "test"
    value = "val2"
  }

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

data "aiven_opensearch" "common-os" {
  service_name = aiven_opensearch.bar-os.service_name
  project      = aiven_opensearch.bar-os.project

  depends_on = [aiven_opensearch.bar-os]
}`, projectName, serviceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest3.TestAccPreCheck(t) },
		ProviderFactories: acctest3.TestAccProviderFactories,
		CheckDestroy:      acctest3.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					acctest3.TestAccCheckAivenServiceCommonAttributes("data.aiven_opensearch.common-os"),
					testAccCheckAivenServiceOSAttributes("data.aiven_opensearch.common-os"),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"),
					resource.TestCheckResourceAttrSet(resourceName, "service_host"),
					resource.TestCheckResourceAttrSet(resourceName, "service_port"),
					resource.TestCheckResourceAttrSet(resourceName, "service_uri"),
				),
			},
			{
				Config:             manifestDoubleTag,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
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
			return fmt.Errorf("expected to get a correct opensearch.public_access.opensearch from Aiven")
		}

		if a["opensearch_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		return nil
	}
}
