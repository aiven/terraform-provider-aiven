//go:build backwardcompat

package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestAccAivenOpenSearchUserBackwardCompatibility verifies that the OpenSearch user resource
// maintains backward compatibility with previous provider versions.
func TestAccAivenOpenSearchUserBackwardCompatibility(t *testing.T) {
	var (
		projectName     = acc.ProjectName()
		rName           = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		serviceName     = fmt.Sprintf("test-os-bc-%s", rName)
		userName        = fmt.Sprintf("user-bc-%s", rName)
		providerVersion = "4.47.0"
	)

	config := fmt.Sprintf(`
resource "aiven_opensearch" "os" {
  project                 = %[1]q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = %[2]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_opensearch_user" "test" {
  project      = aiven_opensearch.os.project
  service_name = aiven_opensearch.os.service_name
  username     = %[3]q
}

data "aiven_opensearch_user" "test" {
  project      = aiven_opensearch_user.test.project
  service_name = aiven_opensearch_user.test.service_name
  username     = aiven_opensearch_user.test.username
}`, projectName, serviceName, userName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: providerVersion,
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("aiven_opensearch.os", "state", "RUNNING"),
				resource.TestCheckResourceAttrSet("aiven_opensearch.os", "id"),

				resource.TestCheckResourceAttrSet("aiven_opensearch_user.test", "id"),
				resource.TestCheckResourceAttr("aiven_opensearch_user.test", "username", userName),
				resource.TestCheckResourceAttrSet("aiven_opensearch_user.test", "password"),
				resource.TestCheckResourceAttr("aiven_opensearch_user.test", "type", "normal"),

				resource.TestCheckResourceAttr("data.aiven_opensearch_user.test", "username", userName),
				resource.TestCheckResourceAttrSet("data.aiven_opensearch_user.test", "password"),
			),
		}),
	})
}
