package project_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_project", &resource.Sweeper{
		Name:         "aiven_project",
		F:            sweepProjects,
		Dependencies: []string{"aiven_service"},
	})
}

func sweepProjects(region string) error {
	client, err := acc.SharedClient(region)
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
	rName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "billing_address"),
					resource.TestCheckResourceAttrSet(resourceName, "billing_extra_text"),
					resource.TestCheckResourceAttrSet(resourceName, "country_code"),
					resource.TestCheckResourceAttrSet(resourceName, "default_cloud"),
					resource.TestCheckResourceAttrSet(resourceName, "billing_currency"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resourceName, "vat_id"),
				),
			},
			{
				Config: testAccProjectCopyFromProjectResource(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectAttributes("data.aiven_project.project"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName2)),
					resource.TestCheckResourceAttrSet(resourceName, "billing_address"),
					resource.TestCheckResourceAttrSet(resourceName, "billing_extra_text"),
					resource.TestCheckResourceAttrSet(resourceName, "country_code"),
					resource.TestCheckResourceAttrSet(resourceName, "default_cloud"),
					resource.TestCheckResourceAttrSet(resourceName, "billing_currency"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_cert"),
					resource.TestCheckResourceAttrSet(resourceName, "vat_id"),
					resource.TestCheckResourceAttrSet(resourceName, "billing_group"),
				),
			},
		},
	})
}

func TestAccAivenProject_accounts(t *testing.T) {
	resourceName := "aiven_project.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
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
		  project            = "test-acc-pr-%s"
		  account_id         = aiven_account.foo.account_id
		  billing_address    = "MXQ4+M5 New York, United States"
		  billing_extra_text = "some extra text ..."
		  country_code       = "DE"
		  default_cloud      = "aws-eu-west-2"
		  billing_currency   = "EUR"
		  vat_id             = "123"
		}
		
		data "aiven_project" "project" {
		  project = aiven_project.foo.project
		
		  depends_on = [aiven_project.foo]
		}`,
		name, name)
}

func testAccProjectResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
		  project            = "test-acc-pr-%s"
		  billing_address    = "MXQ4+M5 New York, United States"
		  billing_extra_text = "some extra text ..."
		  country_code       = "DE"
		  default_cloud      = "aws-eu-west-2"
		  billing_currency   = "EUR"
		  vat_id             = "123"
		}
		
		data "aiven_project" "project" {
		  project    = aiven_project.foo.project
		  depends_on = [aiven_project.foo]
		}`,
		name)
}

func testAccProjectCopyFromProjectResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_billing_group" "foo" {
		  name             = "test-acc-bg-%s"
		  billing_currency = "USD"
		  vat_id           = "123"
		}
		
		resource "aiven_project" "source" {
		  project       = "test-acc-pr-source-%s"
		  billing_group = aiven_billing_group.foo.id
		}
		
		resource "aiven_project" "foo" {
		  project                          = "test-acc-pr-%s"
		  copy_from_project                = aiven_project.source.project
		  use_source_project_billing_group = true
		}
		
		data "aiven_project" "project" {
		  project    = aiven_project.foo.project
		  depends_on = [aiven_project.foo]
		}`,
		name, name, name)
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
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project" {
			continue
		}

		p, err := c.Projects.Get(rs.Primary.ID)
		if err != nil {
			errStatus := err.(aiven.Error).Status
			if errStatus != 404 && errStatus != 403 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("project (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
