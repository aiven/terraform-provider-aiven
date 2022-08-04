package grafana_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAiven_grafana(t *testing.T) {
	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_grafana.common"),
					testAccCheckAivenServiceGrafanaAttributes("data.aiven_grafana.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "grafana"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config:             testAccGrafanaDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func TestAccAiven_grafana_user_config(t *testing.T) {
	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_grafana.common"),
					testAccCheckAivenServiceGrafanaAttributes("data.aiven_grafana.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.ip_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.alerting_enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "grafana_user_config.0.public_access.0.grafana", "true",
					),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_grafana" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  grafana_user_config {
    alerting_enabled = true
    ip_filter        = ["127.0.0.1/32", "10.13.37.0/24"]

    public_access {
      grafana = false
    }
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.ip_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.alerting_enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "grafana_user_config.0.public_access.0.grafana", "false",
					),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_grafana" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  grafana_user_config {
    ip_filter = ["10.13.37.0/24", "127.0.0.1/32"]
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.ip_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.alerting_enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "grafana_user_config.0.public_access.0.grafana", "false",
					),
				),
			},
			{
				Config: fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_grafana" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.ip_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.alerting_enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "grafana_user_config.0.public_access.0.grafana", "false",
					),
				),
			},
		},
	})
}

func testAccGrafanaResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_grafana" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  grafana_user_config {
    ip_filter        = ["0.0.0.0/0"]
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}
func testAccGrafanaDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_grafana" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-sr-%s"
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

  grafana_user_config {
    ip_filter        = ["0.0.0.0/0"]
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}
