package mysql_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

<<<<<<< HEAD
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
=======
	acctest3 "github.com/aiven/terraform-provider-aiven/internal/acctest"
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
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
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
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
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
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

		return nil
	}
}
