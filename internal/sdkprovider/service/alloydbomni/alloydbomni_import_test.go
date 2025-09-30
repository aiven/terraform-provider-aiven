package alloydbomni_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAivenAlloyDBOmni_import(t *testing.T) {
	t.Skip("Deprecated resource")

	resourceName := "aiven_alloydbomni.main"
	projectName := acc.ProjectName()
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniImportResource(projectName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s-main", rName)),
				),
			},
			{
				Config:       testAccAlloyDBOmniImportResource(projectName, rName),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("expected resource '%s' to be present in the state", resourceName)
					}
					return rs.Primary.ID, nil
				},
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					assert := assert.New(t)
					if !assert.Len(s, 1, "expected only one instance to be imported") {
						return fmt.Errorf("state: %#v", s)
					}
					attributes := s[0].Attributes
					assert.Equal(projectName, attributes["project"])
					assert.Equal("google-europe-west1", attributes["cloud_name"])
					assert.Equal("startup-4", attributes["plan"])
					assert.Equal(fmt.Sprintf("test-acc-sr-%s-main", rName), attributes["service_name"])
					assert.Equal("monday", attributes["maintenance_window_dow"])
					assert.Equal("10:00:00", attributes["maintenance_window_time"])
					assert.Equal("30GiB", attributes["additional_disk_space"])
					assert.Equal("alloydbomniimporttest@aiven.io", attributes["tech_emails.0.email"])
					assert.Equal("test-key", attributes["tag.0.key"])
					assert.Equal("test-value", attributes["tag.0.value"])
					assert.Equal(fmt.Sprintf("%s/test-acc-sr-%s-main", projectName, rName), s[0].ID)
					return nil
				},
			},
		},
	})
}

func testAccAlloyDBOmniImportResource(project string, name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "main" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s-main"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  additional_disk_space   = "30GiB"

  tech_emails {
    email = "alloydbomniimporttest@aiven.io"
  }

  tag {
    key   = "test-key"
    value = "test-value"
  }
}

resource "aiven_alloydbomni" "read_replica" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s-read-replica"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  additional_disk_space   = "30GiB"

  service_integrations {
    source_service_name = aiven_alloydbomni.main.service_name
    integration_type    = "read_replica"
  }
}
`, project, name, name)
}
