package project_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenProjectUser_basic(t *testing.T) {
	resourceName := "aiven_project_user.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenProjectUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectUserAttributes("data.aiven_project_user.user"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("ivan.savciuc+%s@aiven.fi", rName)),
					resource.TestCheckResourceAttr(resourceName, "member_type", "admin"),
				),
			},
			{
				Config: testAccProjectUserDeveloperResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectUserAttributes("data.aiven_project_user.user"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("ivan.savciuc+%s@aiven.fi", rName)),
					resource.TestCheckResourceAttr(resourceName, "member_type", "developer"),
				),
			},
		},
	})
}

func testAccCheckAivenProjectUserResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each project is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project_user" {
			continue
		}

		projectName, email, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, i, err := c.ProjectUsers.Get(ctx, projectName, email)
		if err != nil {
			var e aiven.Error
			if errors.As(err, &e) && e.Status != 404 && e.Status != 403 {
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
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_project" "foo" {
  project       = "test-acc-pr-%[1]s"
  default_cloud = "aws-eu-west-2"
  parent_id     = aiven_organization.foo.id
}

resource "aiven_project_user" "bar" {
  project     = aiven_project.foo.project
  email       = "ivan.savciuc+%[1]s@aiven.fi"
  member_type = "admin"
}

data "aiven_project_user" "user" {
  project = aiven_project_user.bar.project
  email   = aiven_project_user.bar.email

  depends_on = [aiven_project_user.bar]
}`, name)
}

func testAccProjectUserDeveloperResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_organization" "foo" {
  name = "test-acc-org-%[1]s"
}

resource "aiven_project" "foo" {
  project       = "test-acc-pr-%[1]s"
  default_cloud = "aws-eu-west-2"
  parent_id     = aiven_organization.foo.id
}

resource "aiven_project_user" "bar" {
  project     = aiven_project.foo.project
  email       = "ivan.savciuc+%[1]s@aiven.fi"
  member_type = "developer"
}

data "aiven_project_user" "user" {
  project = aiven_project_user.bar.project
  email   = aiven_project_user.bar.email

  depends_on = [aiven_project_user.bar]
}`, name)
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
