package dragonfly_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAiven_dragonfly(t *testing.T) {
	resourceName := "aiven_dragonfly.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDragonflyResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_dragonfly.common"),
					testAccCheckAivenServiceDragonflyAttributes("data.aiven_dragonfly.common"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.0.email", "techsupport@company.com"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "dragonfly"),
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
				Config: testAccDragonflyRemoveEmailsResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_dragonfly.common"),
					testAccCheckAivenServiceDragonflyAttributes("data.aiven_dragonfly.common"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.#", "0"),
				),
			},
			{
				Config:             testAccDragonflyDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func testAccDragonflyResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_dragonfly" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  tech_emails {
    email = "techsupport@company.com"
  }

  dragonfly_user_config {
    public_access {
      dragonfly = true
    }
  }
}

data "aiven_dragonfly" "common" {
  service_name = aiven_dragonfly.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_dragonfly.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccDragonflyRemoveEmailsResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_dragonfly" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  dragonfly_user_config {
    public_access {
      dragonfly = true
    }
  }
}

data "aiven_dragonfly" "common" {
  service_name = aiven_dragonfly.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_dragonfly.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccDragonflyDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_dragonfly" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
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

  dragonfly_user_config {
    public_access {
      dragonfly = true
    }
  }
}

data "aiven_dragonfly" "common" {
  service_name = aiven_dragonfly.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_dragonfly.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

// Dragonfly service tests
func TestAccAivenService_dragonfly(t *testing.T) {
	resourceName := "aiven_dragonfly.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDragonflyServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_dragonfly.common"),
					testAccCheckAivenServiceDragonflyAttributes("data.aiven_dragonfly.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "dragonfly"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccDragonflyServiceResourceWithPersistenceOff(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_dragonfly.common"),
					testAccCheckAivenServiceDragonflyAttributes("data.aiven_dragonfly.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "dragonfly"),
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

func testAccDragonflyServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_dragonfly" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  dragonfly_user_config {
    public_access {
      dragonfly = true
    }
  }
}

data "aiven_dragonfly" "common" {
  service_name = aiven_dragonfly.bar.service_name
  project      = aiven_dragonfly.bar.project

  depends_on = [aiven_dragonfly.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccDragonflyServiceResourceWithPersistenceOff(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_dragonfly" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  dragonfly_user_config {
    public_access {
      dragonfly = true
    }
  }
}

data "aiven_dragonfly" "common" {
  service_name = aiven_dragonfly.bar.service_name
  project      = aiven_dragonfly.bar.project

  depends_on = [aiven_dragonfly.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceDragonflyAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "dragonfly" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["dragonfly_user_config.0.public_access.0.dragonfly"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.dragonfly from Aiven")
		}

		if a["dragonfly_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		return nil
	}
}
