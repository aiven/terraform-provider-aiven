package valkey_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
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

func TestAccAivenValkey_WriteOnlyPassword(t *testing.T) {
	resourceName := "aiven_valkey.test"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	password1 := "CustomPassword123!@#"
	password2 := "RotatedPassword456$%^"
	password3 := "FinalPassword789&*("

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source:            "hashicorp/random",
				VersionConstraint: ">= 3.0.0",
			},
		},
		CheckDestroy: acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Create service with auto-generated password
				Config: testAccValkeyResourceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-valkey-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"), // verify password is in state
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo_version"),
					testAccCheckValkeyServicePasswordExists(resourceName),
				),
			},
			{
				// Step 2: Migrate to write-only password with explicit value
				Config: testAccValkeyResourceWriteOnlyExplicit(rName, password1, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-valkey-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"), // verify WO password NOT in state
					resource.TestCheckResourceAttr(resourceName, "service_password", ""),  // verify old password cleared
					testAccCheckValkeyServicePasswordMatches(resourceName, password1),     // verify actual password
				),
			},
			{
				// Step 3: Rotate password by incrementing version
				Config: testAccValkeyResourceWriteOnlyExplicit(rName, password2, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_password", ""),
					testAccCheckValkeyServicePasswordMatches(resourceName, password2), // verify password rotated
				),
			},
			{
				// Step 4: Another rotation
				Config: testAccValkeyResourceWriteOnlyExplicit(rName, password3, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "3"),
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_password", ""),
					testAccCheckValkeyServicePasswordMatches(resourceName, password3),
				),
			},
			{
				// Step 5: Switch back to auto-generated password
				Config: testAccValkeyResourceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-valkey-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"), // password back in state
					resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
					resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "0"), // version reset to 0
					testAccCheckValkeyServicePasswordExists(resourceName),
					// Verify password changed from the custom one
					testAccCheckValkeyServicePasswordNotMatches(resourceName, password3),
				),
			},
		},
	})
}

func testAccValkeyResourceBasic(name string) string {
	return fmt.Sprintf(`
resource "aiven_valkey" "test" {
  project                 = "%[1]s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-valkey-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
`, acc.ProjectName(), name)
}

func testAccValkeyResourceWriteOnlyExplicit(name, password string, version int) string {
	return fmt.Sprintf(`
resource "aiven_valkey" "test" {
  project                     = "%[1]s"
  cloud_name                  = "google-europe-west1"
  plan                        = "startup-4"
  service_name                = "test-acc-valkey-%[2]s"
  maintenance_window_dow      = "monday"
  maintenance_window_time     = "10:00:00"
  service_password_wo         = "%[3]s"
  service_password_wo_version = %[4]d
}
`, acc.ProjectName(), name, password, version)
}

// testAccCheckValkeyServicePassword is a generic helper that validates the Valkey service password using a custom validation function
func testAccCheckValkeyServicePassword(resourceName string, validateFn func(password string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		projectName, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing resource ID: %w", err)
		}

		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			return fmt.Errorf("error getting client: %w", err)
		}

		// Valkey uses "default" as the default username, not "avnadmin"
		user, err := client.ServiceUserGet(context.Background(), projectName, serviceName, "default")
		if err != nil {
			return fmt.Errorf("error getting default user: %w", err)
		}

		if user.Password == "" {
			return fmt.Errorf("default user password is empty in API response")
		}

		return validateFn(user.Password)
	}
}

// testAccCheckValkeyServicePasswordExists verifies that the service password exists and is not empty
func testAccCheckValkeyServicePasswordExists(resourceName string) resource.TestCheckFunc {
	return testAccCheckValkeyServicePassword(resourceName, func(password string) error {
		return nil // Password existence already checked in the generic function
	})
}

// testAccCheckValkeyServicePasswordMatches verifies that the actual password matches the expected value
func testAccCheckValkeyServicePasswordMatches(resourceName, expectedPassword string) resource.TestCheckFunc {
	return testAccCheckValkeyServicePassword(resourceName, func(password string) error {
		if password != expectedPassword {
			return fmt.Errorf("default user password does not match expected value: got %q, want %q", password, expectedPassword)
		}
		return nil
	})
}

// testAccCheckValkeyServicePasswordNotMatches verifies that the password has changed from a previous value
func testAccCheckValkeyServicePasswordNotMatches(resourceName, oldPassword string) resource.TestCheckFunc {
	return testAccCheckValkeyServicePassword(resourceName, func(password string) error {
		if password == oldPassword {
			return fmt.Errorf("default user password still matches old value: %q", oldPassword)
		}
		return nil
	})
}
