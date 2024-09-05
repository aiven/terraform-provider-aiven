package thanos_test

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

func TestAccAiven_thanos(t *testing.T) {
	resourceName := "aiven_thanos.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThanosResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_thanos.common"),
					testAccCheckAivenServiceThanosAttributes("data.aiven_thanos.common"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.0.email", "techsupport@company.com"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "thanos"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "thanos_user_config.0.query_frontend.0.query_range_align_range_with_step", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"),
					resource.TestCheckResourceAttrSet(resourceName, "service_host"),
					resource.TestCheckResourceAttrSet(resourceName, "service_port"),
					resource.TestCheckResourceAttrSet(resourceName, "service_uri"),
				),
			},
			{
				Config: testAccThanosRemoveEmailsResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_thanos.common"),
					testAccCheckAivenServiceThanosAttributes("data.aiven_thanos.common"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "thanos_user_config.0.query_frontend.0.query_range_align_range_with_step", "true"),
				),
			},
			{
				Config:             testAccThanosDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func testAccThanosResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_thanos" "bar" {
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

  thanos_user_config {
    query_frontend {
      query_range_align_range_with_step = false
    }

    public_access {
      query = true
    }
  }
}

data "aiven_thanos" "common" {
  service_name = aiven_thanos.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_thanos.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccThanosRemoveEmailsResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_thanos" "bar" {
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

  thanos_user_config {
    query_frontend {
      query_range_align_range_with_step = true
    }

    public_access {
      query = true
    }
  }
}

data "aiven_thanos" "common" {
  service_name = aiven_thanos.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_thanos.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccThanosDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_thanos" "bar" {
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

  thanos_user_config {
    public_access {
      query = true
    }
  }
}

data "aiven_thanos" "common" {
  service_name = aiven_thanos.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_thanos.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

// Thanos service tests
func TestAccAivenService_thanos(t *testing.T) {
	resourceName := "aiven_thanos.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThanosServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_thanos.common"),
					testAccCheckAivenServiceThanosAttributes("data.aiven_thanos.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "thanos"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccThanosServiceResourceWithPersistenceOff(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_thanos.common"),
					testAccCheckAivenServiceThanosAttributes("data.aiven_thanos.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "thanos"),
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

func testAccThanosServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_thanos" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  thanos_user_config {
    public_access {
      query = true
    }
  }
}

data "aiven_thanos" "common" {
  service_name = aiven_thanos.bar.service_name
  project      = aiven_thanos.bar.project

  depends_on = [aiven_thanos.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccThanosServiceResourceWithPersistenceOff(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_thanos" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  thanos_user_config {
    public_access {
      query = true
    }
  }
}

data "aiven_thanos" "common" {
  service_name = aiven_thanos.bar.service_name
  project      = aiven_thanos.bar.project

  depends_on = [aiven_thanos.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceThanosAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "thanos" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["thanos_user_config.0.public_access.0.query"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.query from Aiven")
		}

		if a["thanos_user_config.0.public_access.0.store"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.store from Aiven")
		}

		if a["thanos.0.uris.#"] == "" {
			return fmt.Errorf("expected to get correct uris from Aiven")
		}

		if a["thanos.0.query_frontend_uri"] == "" {
			return fmt.Errorf("expected to get correct query_frontend_uri from Aiven")
		}

		if a["thanos.0.query_uri"] == "" {
			return fmt.Errorf("expected to get correct query_uri from Aiven")
		}

		if a["thanos.0.receiver_ingesting_remote_write_uri"] == "" {
			return fmt.Errorf("expected to get correct receiver_ingesting_remote_write_uri from Aiven")
		}

		if a["thanos.0.receiver_remote_write_uri"] == "" {
			return fmt.Errorf("expected to get correct receiver_remote_write_uri from Aiven")
		}

		return nil
	}
}
