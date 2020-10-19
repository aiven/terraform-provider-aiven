package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"strings"
	"testing"
)

func init() {
	resource.AddTestSweepers("aiven_project", &resource.Sweeper{
		Name:         "aiven_project",
		F:            sweepProjects,
		Dependencies: []string{"aiven_service"},
	})
}

func sweepProjects(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if strings.Contains(project.Name, "test-acc-pr") {
			if err := conn.Projects.Delete(project.Name); err != nil {
				e := err.(aiven.Error)

				// project not found
				if e.Status == 404 {
					continue
				}

				// project with open balance cannot be destroyed
				if strings.Contains(e.Message, "open balance") && e.Status == 403 {
					log.Printf("[DEBUG] project %s with open balance cannot be destroyed", project.Name)
					continue
				}

				return fmt.Errorf("error destroying project %s during sweep: %s", project.Name, err)
			}
		}
	}

	return nil
}

func TestAccAivenProject_basic(t *testing.T) {
	resourceName := "aiven_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenProject_accounts(t *testing.T) {

	resourceName := "aiven_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceAccounts(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
				),
			},
		},
	})
}

func testAccProjectResourceAccounts(name string) string {
	return fmt.Sprintf(`
		resource "aiven_account" "foo" {
			name = "test-acc-ac-%s"
		} 

		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			account_id = aiven_account.foo.account_id
		}

		data "aiven_project" "project" {
			project = aiven_project.foo.project
		}
		`, name, name)
}

func testAccProjectResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		data "aiven_project" "project" {
			project = aiven_project.foo.project
		}
		`, name)
}

func testAccCheckAivenProjectAttributes(n string, attributes ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] project attributes %v", a)

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["ca_cert"] == "" {
			return fmt.Errorf("expected to get an ca_cert from Aiven")
		}

		for _, attr := range attributes {
			if a[attr] == "" {
				return fmt.Errorf("expected to get an %s from Aiven", attr)
			}
		}

		return nil
	}
}

func testAccCheckAivenProjectResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project" {
			continue
		}

		p, err := c.Projects.Get(rs.Primary.ID)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("porject (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
