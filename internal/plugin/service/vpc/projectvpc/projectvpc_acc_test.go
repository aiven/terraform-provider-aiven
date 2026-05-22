package projectvpc_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
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
					resource.TestCheckResourceAttrSet(resourceName, "project_vpc_id"),
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
					resource.TestCheckResourceAttrSet(resourceName, "project_vpc_id"),
				),
			},
			{
				Config:            testAccProjectVPCResourceGetByID(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAivenProjectVPC_backwardCompat(t *testing.T) {
	resourceName := "aiven_project_vpc.bar"
	project := acc.ProjectName()
	config := testAccProjectVPCResourceOnly(project, "google-europe-west4", "192.168.2.0/24")
	oldChecks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "project", project),
		resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west4"),
		resource.TestCheckResourceAttr(resourceName, "network_cidr", "192.168.2.0/24"),
		resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
	)
	newChecks := resource.ComposeTestCheckFunc(
		oldChecks,
		resource.TestCheckResourceAttrSet(resourceName, "project_vpc_id"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aiven": acc.ExternalAivenProvider(t, "4.56.0"),
				},
				Config: config,
				Check:  oldChecks,
			},
			{
				ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
				Config:                   config,
				ExpectNonEmptyPlan:       false,
				Check:                    newChecks,
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

func testAccProjectVPCResourceOnly(project, cloud, cidr string) string {
	return fmt.Sprintf(`
resource "aiven_project_vpc" "bar" {
  project      = %[1]q
  cloud_name   = %[2]q
  network_cidr = %[3]q
}
`, project, cloud, cidr)
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
	c, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project_vpc" {
			continue
		}

		project, vpcID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		vpc, err := c.VpcGet(ctx, project, vpcID)
		if err != nil {
			var e avngen.Error
			if errors.As(err, &e) && e.Status != http.StatusNotFound && e.Status != http.StatusForbidden {
				return err
			}
		}

		if vpc != nil {
			return fmt.Errorf("project vpc (%s) still exists", rs.Primary.ID)
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
