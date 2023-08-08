package flink_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acctest3 "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAiven_flink tests Flink resource.
func TestAccAiven_flink(t *testing.T) {
	projectName := os.Getenv("AIVEN_PROJECT_NAME")

	randString := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) }

	serviceName := fmt.Sprintf("test-acc-flink-%s", randString())

	manifest := fmt.Sprintf(`
variable "project_name" {
  type    = string
  default = "%s"
}
variable "service_name" {
  type    = string
  default = "%s"
}

resource "aiven_flink" "bar" {
  project                 = var.project_name
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.service_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  flink_user_config {
    number_of_task_slots = 10
  }
}

data "aiven_flink" "service" {
  service_name = aiven_flink.bar.service_name
  project      = aiven_flink.bar.project
}`, projectName,
		serviceName,
	)

	manifestApplication := fmt.Sprintf(`
variable "project_name" {
  type    = string
  default = "%s"
}
variable "service_name" {
  type    = string
  default = "%s"
}

resource "aiven_flink" "bar" {
  project                 = var.project_name
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.service_name
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  flink_user_config {
    number_of_task_slots = 10
  }
}

resource "aiven_flink_application" "foo" {
  project      = var.project_name
  service_name = aiven_flink.bar.service_name
  name         = "test"
}

data "aiven_flink_application" "bar" {
  project      = aiven_flink_application.foo.project
  service_name = aiven_flink_application.foo.service_name
  name         = aiven_flink_application.foo.name
}

data "aiven_flink" "service" {
  service_name = aiven_flink.bar.service_name
  project      = aiven_flink.bar.project
}`, projectName,
		serviceName,
	)

	manifestDoubleTags := fmt.Sprintf(`
variable "project_name" {
  type    = string
  default = "%s"
}
variable "service_name" {
  type    = string
  default = "%s"
}

resource "aiven_flink" "bar" {
  project                 = var.project_name
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = var.service_name
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

  flink_user_config {
    number_of_task_slots = 10
  }
}

resource "aiven_flink_application" "foo" {
  project      = var.project_name
  service_name = aiven_flink.bar.service_name
  name         = "test"
}

data "aiven_flink_application" "bar" {
  project      = aiven_flink_application.foo.project
  service_name = aiven_flink_application.foo.service_name
  name         = aiven_flink_application.foo.name
}

data "aiven_flink" "service" {
  service_name = aiven_flink.bar.service_name
  project      = aiven_flink.bar.project
}`, projectName,
		serviceName,
	)

	resourceName := "aiven_flink.bar"

	resourceNameApplication := "aiven_flink_application.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest3.TestAccPreCheck(t) },
		ProviderFactories: acctest3.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					acctest3.TestAccCheckAivenServiceCommonAttributes("data.aiven_flink.service"),
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "flink"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "flink_user_config.0.number_of_task_slots", "10"),
				),
			},
			{
				Config: manifestApplication,
				Check: resource.ComposeTestCheckFunc(
					aivenFlinkApplicationAttributes("data.aiven_flink_application.bar"),
					resource.TestCheckResourceAttr(resourceNameApplication, "name", "test"),
				),
			},
			{
				Config:             manifestDoubleTags,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func aivenFlinkApplicationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		if rs.Primary.Attributes["project"] == "" {
			return fmt.Errorf("no project name is set")
		}

		if rs.Primary.Attributes["service_name"] == "" {
			return fmt.Errorf("no service name is set")
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("no application name is set")
		}

		if rs.Primary.Attributes["application_id"] == "" {
			return fmt.Errorf("no application id is set")
		}

		if rs.Primary.Attributes["created_at"] == "" {
			return fmt.Errorf("no created at is set")
		}

		if rs.Primary.Attributes["created_by"] == "" {
			return fmt.Errorf("no created by is set")
		}

		if rs.Primary.Attributes["updated_at"] == "" {
			return fmt.Errorf("no updated at is set")
		}

		if rs.Primary.Attributes["updated_by"] == "" {
			return fmt.Errorf("no updated by is set")
		}

		return nil
	}
}
