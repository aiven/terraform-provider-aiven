//go:build backwardcompat

package mysql_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAivenMySQLUserBackwardCompatibility verifies that the MySQL user resource
// maintains backward compatibility with previous provider versions.
func TestAccAivenMySQLUserBackwardCompatibility(t *testing.T) {
	var (
		projectName = acc.ProjectName()
		rName       = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName = fmt.Sprintf("test-mysql-bc-%s", rName)
		userName    = fmt.Sprintf("user-bc-%s", rName)
	)

	config := fmt.Sprintf(`
resource "aiven_mysql" "mysql" {
  project                 = %[1]q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = %[2]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_mysql_user" "test" {
  project      = aiven_mysql.mysql.project
  service_name = aiven_mysql.mysql.service_name
  username     = %[3]q
}

data "aiven_mysql_user" "test" {
  project      = aiven_mysql_user.test.project
  service_name = aiven_mysql_user.test.service_name
  username     = aiven_mysql_user.test.username
}`, projectName, serviceName, userName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.47.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("aiven_mysql.mysql", "state", "RUNNING"),
				resource.TestCheckResourceAttrSet("aiven_mysql.mysql", "id"),

				resource.TestCheckResourceAttrSet("aiven_mysql_user.test", "id"),
				resource.TestCheckResourceAttr("aiven_mysql_user.test", "username", userName),
				resource.TestCheckResourceAttrSet("aiven_mysql_user.test", "password"),
				resource.TestCheckResourceAttr("aiven_mysql_user.test", "type", "normal"),

				resource.TestCheckResourceAttr("data.aiven_mysql_user.test", "username", userName),
				resource.TestCheckResourceAttrSet("data.aiven_mysql_user.test", "password"),
			),
		}),
	})
}
