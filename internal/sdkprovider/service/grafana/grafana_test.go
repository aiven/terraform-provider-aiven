package grafana_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/converters"
)

func TestAccAiven_grafana(t *testing.T) {
	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
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
	// Parallel tests share os env, can't isolate this one
	os.Setenv(converters.AllowIPFilterPurge, "1")

	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_grafana.common"),
					testAccCheckAivenServiceGrafanaAttributes("data.aiven_grafana.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.ip_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.alerting_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.public_access.0.grafana", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.public_access.0.grafana", "false"),
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
    // Hides array, gets 0 length
    // ip_filter        = ["127.0.0.1/32", "10.13.37.0/24"]
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
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.public_access.0.grafana", "false"),
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
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.public_access.0.grafana", "false"),
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
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.public_access.0.grafana", "false"),
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

// Grafana service tests
func TestAccAivenService_grafana(t *testing.T) {
	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaServiceResource(rName),
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
				Config: testAccGrafanaServiceCustomIPFiltersResource(rName2),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_grafana.common"),
					testAccCheckAivenServiceGrafanaAttributes("data.aiven_grafana.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName2)),
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
		},
	})
}

func testAccGrafanaServiceResource(name string) string {
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
  project      = aiven_grafana.bar.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccGrafanaServiceCustomIPFiltersResource(name string) string {
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

  grafana_user_config {
    ip_filter        = ["130.27.80.0/24"]
    alerting_enabled = true

    public_access {
      grafana = true
    }
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = aiven_grafana.bar.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceGrafanaAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "grafana" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["grafana_user_config.0.alerting_enabled"] != "true" {
			return fmt.Errorf("expected to get an alerting_enabled from Aiven")
		}

		if a["grafana_user_config.0.public_access.0.grafana"] != "true" {
			return fmt.Errorf("expected to get an public_access.grafana from Aiven")
		}

		if a["grafana.0.uris.#"] == "" {
			return fmt.Errorf("expected to get correct uris from Aiven")
		}

		return nil
	}
}

func TestAccAivenService_grafana_with_ip_filter_objects(t *testing.T) {
	t.Skip(`activate when provider can manage default object values`)

	project := os.Getenv("AIVEN_PROJECT_NAME")
	prefix := "test-acc-" + acctest.RandString(7)
	resourceName := "aiven_grafana.grafana"

	// This checks prove that deleting ip_filter_objects does not change the state
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
		resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.ip_filter_object.#", "1"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaServiceResourceWithIPFilterObjects(prefix, project, true),
				Check:  checks,
			},
			{
				Config: testAccGrafanaServiceResourceWithIPFilterObjects(prefix, project, false),
				Check:  checks,
			},
		},
	})
}

//nolint:unused
func testAccGrafanaServiceResourceWithIPFilterObjects(prefix, project string, addIPFilterObjs bool) string {
	ipFilterObjs := `ip_filter_object {
      description = "test"
      network     = "1.3.3.7/32"
    }`

	if !addIPFilterObjs {
		ipFilterObjs = ``
	}

	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

resource "aiven_grafana" "grafana" {
  project                 = data.aiven_project.project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "%[1]s-grafana"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  grafana_user_config {
    alerting_enabled = true
    %[3]s

    public_access {
      grafana = false
    }
  }
}

`, prefix, project, ipFilterObjs)
}

// TestAccAiven_grafana_set_change tests that changing a set actually changes it count
// This is a test for diff suppressor doesn't suppress set's items.
func TestAccAiven_grafana_set_change(t *testing.T) {
	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrafanaResourceSetChange(rName, "100, 101, 111"),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_grafana.common"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.auth_github.0.client_id", "my_client_id"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.auth_github.0.client_secret", "my_client_secret"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.auth_github.0.team_ids.#", "3"),
				),
			},
			{
				Config: testAccGrafanaResourceSetChange(rName, "111"),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_grafana.common"),
					resource.TestCheckResourceAttr(resourceName, "grafana_user_config.0.auth_github.0.team_ids.#", "1"),
				),
			},
		},
	})
}

func testAccGrafanaResourceSetChange(name, teamIDs string) string {
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

  grafana_user_config {
    auth_github {
      client_id     = "my_client_id"
      client_secret = "my_client_secret"
      team_ids      = [%s]
    }
  }
}

data "aiven_grafana" "common" {
  service_name = aiven_grafana.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_grafana.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, teamIDs)
}

func TestAccAiven_grafana_user_config_ip_filter_object(t *testing.T) {
	// Parallel tests share os env, can't isolate this one
	os.Setenv(converters.AllowIPFilterPurge, "1")

	resourceName := "aiven_grafana.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	ipCount := "grafana_user_config.0.ip_filter_object.#"
	ipValue := "grafana_user_config.0.ip_filter_object.*"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				// Creates a service with one ip_filter_object
				Config: testAccAivenGrafanaUserConfigIPFilterObject(rName, `["10.0.0.0/8"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, ipCount, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						ipValue,
						map[string]string{"network": "10.0.0.0/8", "description": "my 10.0.0.0/8"},
					),
				),
			},
			{
				// Adds two more
				Config: testAccAivenGrafanaUserConfigIPFilterObject(rName, `["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, ipCount, "3"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						ipValue,
						map[string]string{
							"network":     "10.0.0.0/8",
							"description": "my 10.0.0.0/8",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						ipValue,
						map[string]string{
							"network":     "172.16.0.0/12",
							"description": "my 172.16.0.0/12",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						ipValue,
						map[string]string{
							"network":     "192.168.0.0/16",
							"description": "my 192.168.0.0/16",
						},
					),
				),
			},
			{
				// Removes one
				Config: testAccAivenGrafanaUserConfigIPFilterObject(rName, `["10.0.0.0/8", "192.168.0.0/16"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, ipCount, "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						ipValue,
						map[string]string{
							"network":     "10.0.0.0/8",
							"description": "my 10.0.0.0/8",
						},
					),
					resource.TestCheckResourceAttr(resourceName, ipCount, "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName,
						ipValue,
						map[string]string{
							"network":     "192.168.0.0/16",
							"description": "my 192.168.0.0/16",
						},
					),
				),
			},
			{
				// Reorders ip_filter_objects, but the plan is empty because it is set type
				Config:   testAccAivenGrafanaUserConfigIPFilterObject(rName, `["192.168.0.0/16", "10.0.0.0/8"]`),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, ipCount, "2"),
				),
			},
			{
				// Removes entries
				Config: testAccAivenGrafanaUserConfigIPFilterObject(rName, `[]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, ipCount, "0"),
				),
			},
		},
	})
}

func testAccAivenGrafanaUserConfigIPFilterObject(name, networks string) string {
	return fmt.Sprintf(`
variable "networks" {
  type    = list(string)
  default = %s
}

resource "aiven_grafana" "bar" {
  project                 = %q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-1"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  grafana_user_config {
    dynamic "ip_filter_object" {
      for_each = var.networks
      content {
        network     = ip_filter_object.value
        description = "my ${ip_filter_object.value}"
      }
    }
  }
}
`, networks, os.Getenv("AIVEN_PROJECT_NAME"), name)
}
