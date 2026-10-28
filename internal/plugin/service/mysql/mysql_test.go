package mysql_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAiven_mysql(t *testing.T) {
	resourceName := "aiven_mysql.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMysqlResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_mysql.common"),
					testAccCheckAivenServiceMysqlAttributes("data.aiven_mysql.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "mysql"),
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
				Config:             testAccMysqlDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func testAccMysqlResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_mysql" "bar" {
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

  mysql_user_config {
    mysql {
      sql_mode                = "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE"
      sql_require_primary_key = true
    }

    public_access {
      mysql = true
    }
  }
}

data "aiven_mysql" "common" {
  service_name = aiven_mysql.bar.service_name
  project      = aiven_mysql.bar.project

  depends_on = [aiven_mysql.bar]
}`, acc.ProjectName(), name)
}

func testAccMysqlDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_mysql" "bar" {
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

  mysql_user_config {
    mysql {
      sql_mode                = "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE"
      sql_require_primary_key = true
    }

    public_access {
      mysql = true
    }
  }
}

data "aiven_mysql" "common" {
  service_name = aiven_mysql.bar.service_name
  project      = aiven_mysql.bar.project

  depends_on = [aiven_mysql.bar]
}`, acc.ProjectName(), name)
}

// MySQL service tests
func TestAccAivenService_mysql(t *testing.T) {
	resourceName := "aiven_mysql.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMysqlServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_mysql.common"),
					testAccCheckAivenServiceMysqlAttributes("data.aiven_mysql.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "mysql"),
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

func testAccMysqlServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_mysql" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  mysql_user_config {
    mysql {
      sql_mode                = "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE"
      sql_require_primary_key = true
    }

    public_access {
      mysql = true
    }
  }
}

data "aiven_mysql" "common" {
  service_name = aiven_mysql.bar.service_name
  project      = aiven_mysql.bar.project

  depends_on = [aiven_mysql.bar]
}`, acc.ProjectName(), name)
}

func testAccCheckAivenServiceMysqlAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["mysql_user_config.0.mysql.0.sql_mode"] != "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE" {
			return fmt.Errorf("expected to get a correct sql_mode from Aiven")
		}

		if a["mysql_user_config.0.public_access.0.mysql"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.mysql from Aiven")
		}

		if a["mysql_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		if a["mysql.0.uris.#"] == "" {
			return fmt.Errorf("expected to get correct uris from Aiven")
		}

		if a["mysql.0.params.#"] == "" {
			return fmt.Errorf("expected to get correct params from Aiven")
		}

		if a["mysql.0.params.0.host"] == "" {
			return fmt.Errorf("expected to get correct host from Aiven")
		}

		if a["mysql.0.params.0.port"] == "" {
			return fmt.Errorf("expected to get correct port from Aiven")
		}

		if a["mysql.0.params.0.sslmode"] == "" {
			return fmt.Errorf("expected to get correct sslmode from Aiven")
		}

		if a["mysql.0.params.0.user"] != "avnadmin" {
			return fmt.Errorf("expected to get correct user from Aiven")
		}

		if a["mysql.0.params.0.password"] == "" {
			return fmt.Errorf("expected to get correct password from Aiven")
		}

		if a["mysql.0.params.0.database_name"] != "defaultdb" {
			return fmt.Errorf("expected to get correct database_name from Aiven")
		}

		return nil
	}
}

func TestAccAivenMySQLPasswordRotation(t *testing.T) {
	acc.TestAccCheckAivenServiceWriteOnlyPassword(t, acc.ServicePasswordTestOptions{
		ResourceType: "aiven_mysql",
		Username:     "avnadmin",
	})
}

// TestAccAivenMySQL_AdminPasswordConflicts tests conflicts between service_password_wo and admin_password
func TestAccAivenMySQL_AdminPasswordConflicts(t *testing.T) {
	t.Skip("This test will be enabled once service support write-only passwords")

	resourceName := "aiven_mysql.test"

	// create a service with only admin_password
	t.Run("direct conflict at creation", func(t *testing.T) {
		serviceName := acc.RandName("mysql")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccMySQLResourceWithBothPasswords(serviceName, "custompass123!", 1),
					ExpectError: regexp.MustCompile(
						`cannot set 'service_password_wo' and 'mysql_user_config\.0\.admin_password' simultaneously`,
					),
				},
			},
		})
	})

	// create a service with only admin_password
	t.Run("admin_password works independently", func(t *testing.T) {
		serviceName := acc.RandName("mysql")
		adminPassword := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum) + "_Aa1"
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccMySQLResourceWithAdminPassword(serviceName, adminPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "mysql_user_config.0.admin_password", adminPassword),
						resource.TestCheckResourceAttrSet(resourceName, "service_password"),
						resource.TestCheckNoResourceAttr(resourceName, "service_password_wo_version"),
					),
				},
			},
		})
	})

	// migrate from admin_password to write-only password
	t.Run("migration path: admin_password to write-only", func(t *testing.T) {
		serviceName := acc.RandName("mysql")
		adminPassword := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum) + "_Aa1"
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccMySQLResourceWithAdminPassword(serviceName, adminPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
						resource.TestCheckResourceAttr(resourceName, "mysql_user_config.0.admin_password", adminPassword),
						resource.TestCheckResourceAttrSet(resourceName, "service_password"),
						resource.TestCheckNoResourceAttr(resourceName, "service_password_wo_version"),
					),
				},
				{
					Config: testAccMySQLResourceWithBothPasswordsAfterCreation(serviceName, adminPassword, "newpass123!", 1),
					ExpectError: regexp.MustCompile(
						`cannot set 'service_password_wo' and 'mysql_user_config\.0\.admin_password' simultaneously`,
					),
				},
				{
					Config: testAccMySQLResourceWithWriteOnlyPasswordOnly(serviceName, "newpass123!", 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "1"),
						resource.TestCheckNoResourceAttr(resourceName, "service_password_wo"),
						resource.TestCheckResourceAttr(resourceName, "service_password", ""),
					),
				},
			},
		})
	})

	// admin password cannot be added after creation
	t.Run("reverse migration blocked: write-only to admin_password", func(t *testing.T) {
		serviceName := acc.RandName("mysql")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 func() { acc.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccMySQLResourceWithWriteOnlyPasswordOnly(serviceName, "testpass123!", 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "service_password_wo_version", "1"),
						resource.TestCheckResourceAttr(resourceName, "service_password", ""),
					),
				},
				{
					Config:      testAccMySQLResourceWithBothPasswordsAfterCreation(serviceName, "adminpass123!", "writeonly123!", 1),
					ExpectError: regexp.MustCompile(`cannot set 'service_password_wo' and 'mysql_user_config\.0\.admin_password' simultaneously`),
				},
			},
		})
	})
}

func testAccMySQLResourceWithBothPasswords(name, woPassword string, version int) string {
	return fmt.Sprintf(`
resource "aiven_mysql" "test" {
  project      = %[1]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = %[2]q

  service_password_wo         = %[3]q
  service_password_wo_version = %[4]d

  mysql_user_config {
    admin_password = "ConflictingPassword123!"
  }
}
`, acc.ProjectName(), name, woPassword, version)
}

func testAccMySQLResourceWithAdminPassword(name, adminPassword string) string {
	return fmt.Sprintf(`
resource "aiven_mysql" "test" {
  project      = %[1]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = %[2]q

  mysql_user_config {
    admin_password = %[3]q
  }
}
`, acc.ProjectName(), name, adminPassword)
}

func testAccMySQLResourceWithBothPasswordsAfterCreation(name, adminPassword, woPassword string, version int) string {
	return fmt.Sprintf(`
resource "aiven_mysql" "test" {
  project      = %[1]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = %[2]q

  service_password_wo         = %[4]q
  service_password_wo_version = %[5]d

  mysql_user_config {
    admin_password = %[3]q
  }
}
`, acc.ProjectName(), name, adminPassword, woPassword, version)
}

func testAccMySQLResourceWithWriteOnlyPasswordOnly(name, woPassword string, version int) string {
	return fmt.Sprintf(`
resource "aiven_mysql" "test" {
  project      = %[1]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = %[2]q

  service_password_wo         = %[3]q
  service_password_wo_version = %[4]d
}
`, acc.ProjectName(), name, woPassword, version)
}
