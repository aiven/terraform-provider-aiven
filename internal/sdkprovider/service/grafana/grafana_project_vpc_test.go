package grafana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// TestAccAivenGrafana_projectVPCIDCanBeRemoved tests that `project_vpc_id` can be removed from a service.
// https://github.com/aiven/terraform-provider-aiven/issues/2242
func TestAccAivenGrafana_projectVPCIDCanBeRemoved(t *testing.T) {
	ctx := context.Background()
	project := acc.ProjectName()
	resourceName := "aiven_grafana.foo"
	serviceName := "test-acc-" + acc.RandStr()
	cloudName := "google-europe-west12"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				// Creates a service with a VPC
				Config: fmt.Sprintf(`
resource "aiven_project_vpc" "foo" {
  project      = %[1]q
  cloud_name   = %[2]q
  network_cidr = "192.168.1.0/24"
}

resource "aiven_grafana" "foo" {
  project                 = aiven_project_vpc.foo.project
  cloud_name              = aiven_project_vpc.foo.cloud_name
  project_vpc_id          = aiven_project_vpc.foo.id
  plan                    = "startup-1"
  service_name            = %[3]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
`, project, cloudName, serviceName),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						client, err := acc.GetTestGenAivenClient()
						if err != nil {
							return err
						}

						service, err := client.ServiceGet(ctx, project, serviceName)
						if err != nil {
							return err
						}

						// Should have a project VPC ID
						if service.ProjectVpcId == "" {
							return fmt.Errorf("expected service %q to have a project VPC ID", serviceName)
						}

						// Check that the project VPC ID is set
						projectVPC := schemautil.BuildResourceID(project, service.ProjectVpcId)
						return resource.TestCheckResourceAttr(resourceName, "project_vpc_id", projectVPC)(s)
					},
				),
			},
			{
				// Removes the `project_vpc_id` field from the service
				Config: fmt.Sprintf(`
resource "aiven_project_vpc" "foo" {
  project      = %[1]q
  cloud_name   = %[2]q
  network_cidr = "192.168.1.0/24"
}

resource "aiven_grafana" "foo" {
  project                 = aiven_project_vpc.foo.project
  cloud_name              = aiven_project_vpc.foo.cloud_name
  plan                    = "startup-1"
  service_name            = %[3]q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}
`, project, cloudName, serviceName),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						client, err := acc.GetTestGenAivenClient()
						if err != nil {
							return err
						}

						service, err := client.ServiceGet(ctx, project, serviceName)
						if err != nil {
							return err
						}

						// Must be empty
						if service.ProjectVpcId != "" {
							return fmt.Errorf("expected service %q to not have a project VPC ID", serviceName)
						}

						// Not set in TF either (not computed).
						// Terraform can't distinguish between "not set" and "empty string",
						// so we check for an empty string here.
						return resource.TestCheckResourceAttr(resourceName, "project_vpc_id", "")(s)
					},
				),
			},
		},
	})
}
