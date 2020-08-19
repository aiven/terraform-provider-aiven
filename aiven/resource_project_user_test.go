package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccAivenProjectUser_basic(t *testing.T) {
	resourceName := "aiven_project_user.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenProjectUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectUserAttributes("data.aiven_project_user.user"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("savciuci+%s@aiven.fi", rName)),
					resource.TestCheckResourceAttr(resourceName, "member_type", "admin"),
				),
			},
		},
	})
}

func testAccCheckAivenProjectUserResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project_user" {
			continue
		}

		projectName, email := splitResourceID2(rs.Primary.ID)
		p, i, err := c.ProjectUsers.Get(projectName, email)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("porject user (%s) still exists", rs.Primary.ID)
		}

		if i != nil {
			return fmt.Errorf("porject user invitation (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccProjectUserResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		resource "aiven_project_user" "bar" {
			project = aiven_project.foo.project
			email = "savciuci+%s@aiven.fi"
			member_type = "admin"
		}

		data "aiven_project_user" "user" {
			project = aiven_project_user.bar.project
			email = aiven_project_user.bar.email
		}
		`, name, name)
}

func testAccCheckAivenProjectUserAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["email"] == "" {
			return fmt.Errorf("expected to get an project user email from Aiven")
		}

		if a["member_type"] == "" {
			return fmt.Errorf("expected to get an project user member_type from Aiven")
		}

		return nil
	}
}
