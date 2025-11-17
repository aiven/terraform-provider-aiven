//go:build backwardcompat

package clickhouse_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAivenClickhouseUserBackwardCompatibility verifies that the ClickHouse user resource
// maintains backward compatibility with previous provider versions.
func TestAccAivenClickhouseUserBackwardCompatibility(t *testing.T) {
	var (
		projectName = acc.ProjectName()
		rName       = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName = fmt.Sprintf("test-ch-bc-%s", rName)
		userName    = fmt.Sprintf("user-bc-%s", rName)
	)

	config := fmt.Sprintf(`
resource "aiven_clickhouse" "ch" {
  project                 = %[1]q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-16"
  service_name            = %[2]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_user" "test" {
  project      = aiven_clickhouse.ch.project
  service_name = aiven_clickhouse.ch.service_name
  username     = %[3]q
}

data "aiven_clickhouse_user" "test" {
  project      = aiven_clickhouse_user.test.project
  service_name = aiven_clickhouse_user.test.service_name
  username     = aiven_clickhouse_user.test.username
}`, projectName, serviceName, userName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.47.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("aiven_clickhouse.ch", "state", "RUNNING"),
				resource.TestCheckResourceAttrSet("aiven_clickhouse.ch", "id"),

				resource.TestCheckResourceAttrSet("aiven_clickhouse_user.test", "id"),
				resource.TestCheckResourceAttr("aiven_clickhouse_user.test", "username", userName),
				resource.TestCheckResourceAttrSet("aiven_clickhouse_user.test", "password"),
				resource.TestCheckResourceAttrSet("aiven_clickhouse_user.test", "uuid"),
				resource.TestCheckResourceAttrSet("aiven_clickhouse_user.test", "required"),

				resource.TestCheckResourceAttr("data.aiven_clickhouse_user.test", "username", userName),
				resource.TestCheckResourceAttrSet("data.aiven_clickhouse_user.test", "uuid"),
			),
		}),
	})
}
