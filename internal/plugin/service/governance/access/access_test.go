package access_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

// TestGovernanceAccess tests the aiven_governance_access resource.
func TestGovernanceAccess(t *testing.T) {
	projectName := acc.ProjectName()
	organizationName := acc.OrganizationName()
	userGroupName := acc.RandName("governance-ug")
	serviceName := acc.RandName("governance-ser")
	resourceName := acc.RandName("governance-res")
	tfResourceID := "aiven_governance_access.foo"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testGovernanceAccessConfig(
					organizationName,
					userGroupName,
					projectName,
					serviceName,
					resourceName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfResourceID, "organization_id"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_name"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_type"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.project"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.service_name"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.username"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.acls.0.resource_name"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.acls.0.resource_type"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.acls.0.operation"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.acls.0.permission_type"),
					resource.TestCheckResourceAttrSet(tfResourceID, "access_data.0.acls.0.host"),
				),
			},
		},
	})
}

func testGovernanceAccessConfig(
	organizationName string,
	userGroupName string,
	projectName string,
	serviceName string,
	resourceName string,
) string {
	return fmt.Sprintf(`
data "aiven_organization" "foo" {
  name = "%s"
}

resource "aiven_organization_user_group" "foo" {
  name            = "test-acc-u-grp-%s"
  organization_id = data.aiven_organization.foo.id
  description     = "test"
}

data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "foo" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_kafka_topic" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.foo.service_name
  topic_name   = "test-acc-topic-%s"
  partitions   = 3
  replication  = 2
}

resource "aiven_governance_access" "foo" {
  organization_id     = data.aiven_organization.foo.id
  access_name         = "My access"
  access_type         = "KAFKA"
  owner_user_group_id = aiven_organization_user_group.foo.group_id

  access_data {
    project      = data.aiven_project.foo.project
    service_name = aiven_kafka.foo.service_name

    acls {
      resource_name   = "foo"
      resource_type   = "Topic"
      operation       = "Write"
      permission_type = "ALLOW"
      host            = "*"
    }

    acls {
      resource_name   = "bar"
      resource_type   = "Topic"
      operation       = "Read"
      permission_type = "ALLOW"
    }
  }
}`, organizationName, userGroupName, projectName, serviceName, resourceName)
}
