package vpc_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client"
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
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenProjectVPCResourceDestroy,
		Steps: []resource.TestStep{
			{
				PlanOnly:    true,
				Config:      testAccProjectVPCResourceFail(rName),
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
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west2"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
				),
			},
			{
				Config: testAccProjectVPCResourceGetByID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectVPCAttributes("data.aiven_project_vpc.vpc2"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
}`, os.Getenv("AIVEN_PROJECT_NAME"))
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
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each project VPC is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project_vpc" {
			continue
		}

		projectName, vpcID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		vpc, err := c.VPCs.Get(projectName, vpcID)
		if err != nil {
			errStatus := err.(aiven.Error).Status
			if errStatus != 404 && errStatus != 403 {
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
`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccProjectVPCResourceFail(name string) string {
	return fmt.Sprintf(`
resource "aiven_project" "foo" {
  project = "test-acc-pr-%s"
}

resource "aiven_project_vpc" "bar" {
  project      = aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  network_cidr = "192.168.0.0/24"
}

data "aiven_project_vpc" "vpc" {
  vpc_id = "some_wrong_id"
}`, name)
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
}`, os.Getenv("AIVEN_PROJECT_NAME"))
}
