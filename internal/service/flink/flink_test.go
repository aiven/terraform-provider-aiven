package flink_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
  plan                    = "startup-4"
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
  plan                    = "startup-4"
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
  plan                    = "startup-4"
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

data "aiven_flink" "service" {
  service_name = aiven_flink.bar.service_name
  project      = aiven_flink.bar.project
}`, projectName,
		serviceName,
	)

	resourceName := "aiven_flink.bar"

	resourceNameApplication := "aiven_flink_application.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: manifest,
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_flink.service"),
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
