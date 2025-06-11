package vpc_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenProjectVPC_basic(t *testing.T) {
	resourceName := "aiven_project_vpc.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenProjectVPCResourceDestroy,
		Steps: []resource.TestStep{
			{
				PlanOnly:    true,
				Config:      testAccProjectVPCResourceFail(),
				ExpectError: regexp.MustCompile("invalid resource id"),
			},
			{
				Config:      testAccServiceProjectVPCResourceFail(rName),
				ExpectError: regexp.MustCompile("invalid project_vpc_id"),
			},
			{
				Config: testAccProjectVPCResource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectVPCAttributes("data.aiven_project_vpc.vpc"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west2"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
				),
			},
			{
				Config: testAccProjectVPCResourceGetByID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectVPCAttributes("data.aiven_project_vpc.vpc2"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "azure-westeurope"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "192.168.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
				),
			},
		},
	})
}

func testAccProjectVPCResource() string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_project_vpc" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west2"
  network_cidr = "192.168.0.0/24"
}

data "aiven_project_vpc" "vpc" {
  project    = aiven_project_vpc.bar.project
  cloud_name = aiven_project_vpc.bar.cloud_name
}`, acc.ProjectName())
}

func testAccCheckAivenProjectVPCAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["cloud_name"] == "" {
			return fmt.Errorf("expected to get a cloud_name from Aiven")
		}

		if a["network_cidr"] == "" {
			return fmt.Errorf("expected to get a network_cidr from Aiven")
		}

		if a["state"] == "" {
			return fmt.Errorf("expected to get a state from Aiven")
		}

		if a["id"] == "" {
			return fmt.Errorf("expected to get an id from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenProjectVPCResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each project VPC is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project_vpc" {
			continue
		}

		projectName, vpcID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		vpc, err := c.VPCs.Get(ctx, projectName, vpcID)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 && e.Status != 403 {
				return err
			}
		}

		if vpc != nil {
			return fmt.Errorf("porject vpc (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccServiceProjectVPCResourceFail(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project        = data.aiven_project.foo.project
  cloud_name     = "google-europe-west1"
  plan           = "startup-4"
  service_name   = "test-acc-sr-%s"
  project_vpc_id = "wrong_vpc_id"
}
`, acc.ProjectName(), name)
}

func testAccProjectVPCResourceFail() string {
	return `
data "aiven_project_vpc" "vpc" {
  vpc_id = "some_wrong_id"
}`
}

func testAccProjectVPCResourceGetByID() string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_project_vpc" "bar" {
  project      = data.aiven_project.foo.project
  cloud_name   = "azure-westeurope"
  network_cidr = "192.168.1.0/24"
}

data "aiven_project_vpc" "vpc2" {
  vpc_id = aiven_project_vpc.bar.id
}`, acc.ProjectName())
}

// TestAccAivenProjectVPC_can_be_removed tests that `project_vpc_id` can be removed from a service
// https://github.com/aiven/terraform-provider-aiven/issues/2242
func TestAccAivenProjectVPC_can_be_removed(t *testing.T) {
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
