package valkey_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAiven_valkey(t *testing.T) {
	resourceName := "aiven_valkey.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccValkeyResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_valkey.common"),
					testAccCheckAivenServiceValkeyAttributes("data.aiven_valkey.common"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.0.email", "techsupport@company.com"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "valkey"),
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
				Config: testAccValkeyRemoveEmailsResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_valkey.common"),
					testAccCheckAivenServiceValkeyAttributes("data.aiven_valkey.common"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_emails.#", "0"),
				),
			},
			{
				Config: testAccValkeyServiceResourceWithPersistenceOff(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_valkey.common"),
					testAccCheckAivenServiceValkeyAttributes("data.aiven_valkey.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "valkey"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config:             testAccValkeyDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func testAccValkeyResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_valkey" "bar" {
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

  valkey_user_config {
    valkey_maxmemory_policy = "allkeys-random"

    public_access {
      valkey = true
    }
  }
}

data "aiven_valkey" "common" {
  service_name = aiven_valkey.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_valkey.bar]
}`, acc.ProjectName(), name)
}

func testAccValkeyRemoveEmailsResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_valkey" "bar" {
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

  valkey_user_config {
    valkey_maxmemory_policy = "allkeys-random"

    public_access {
      valkey = true
    }
  }
}

data "aiven_valkey" "common" {
  service_name = aiven_valkey.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_valkey.bar]
}`, acc.ProjectName(), name)
}

func testAccValkeyServiceResourceWithPersistenceOff(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_valkey" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  valkey_user_config {
    valkey_persistence      = "off"
    valkey_maxmemory_policy = "allkeys-random"

    public_access {
      valkey = true
    }
  }
}

data "aiven_valkey" "common" {
  service_name = aiven_valkey.bar.service_name
  project      = aiven_valkey.bar.project

  depends_on = [aiven_valkey.bar]
}`, acc.ProjectName(), name)
}

func testAccValkeyDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_valkey" "bar" {
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

  valkey_user_config {
    valkey_maxmemory_policy = "allkeys-random"

    public_access {
      valkey = true
    }
  }
}

data "aiven_valkey" "common" {
  service_name = aiven_valkey.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_valkey.bar]
}`, acc.ProjectName(), name)
}

func testAccCheckAivenServiceValkeyAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "valkey" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["valkey_user_config.0.valkey_maxmemory_policy"] != "allkeys-random" {
			return fmt.Errorf("expected to get a correct valkey_maxmemory_policy from Aiven")
		}

		if a["valkey_user_config.0.public_access.0.valkey"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.valkey from Aiven")
		}

		if a["valkey_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		if a["valkey.0.uris.#"] == "" {
			return fmt.Errorf("expected to get correct uris from Aiven")
		}

		if a["valkey.0.slave_uris.#"] == "" {
			return fmt.Errorf("expected to get correct slave_uris from Aiven")
		}

		if a["valkey.0.replica_uri"] != "" {
			return fmt.Errorf("expected to get correct replica_uri from Aiven")
		}

		if a["valkey.0.password"] == "" {
			return fmt.Errorf("expected to get correct password from Aiven")
		}

		return nil
	}
}
