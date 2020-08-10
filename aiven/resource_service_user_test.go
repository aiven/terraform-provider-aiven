package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"testing"
)

func TestAccAivenServiceUser_basic(t *testing.T) {
	resourceName := "aiven_service_user.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenServiceUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceUserAttributes("data.aiven_service_user.user"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "username", fmt.Sprintf("user-%s", rName)),
				),
			},
		},
	})
}

func testAccCheckAivenServiceUserResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_service is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_user" {
			continue
		}

		projectName, serviceName, username := splitResourceID3(rs.Primary.ID)
		p, err := c.ServiceUsers.Get(projectName, serviceName, username)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("service user (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccServiceUserResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
			project = "%s"
		}
		
		resource "aiven_service" "bar" {
			project = data.aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "pg"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			pg_user_config {
				pg_version = 11
			}
		}
		
		resource "aiven_service_user" "foo" {
			service_name = aiven_service.bar.service_name
			project = data.aiven_project.foo.project
			username = "user-%s"
		}
		
		data "aiven_service_user" "user" {
			service_name = aiven_service.bar.service_name
			project = aiven_service.bar.project
			username = aiven_service_user.foo.username
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccCheckAivenServiceUserAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		log.Printf("[DEBUG] user service attributes %v", a)

		if a["username"] == "" {
			return fmt.Errorf("expected to get a service user username from Aiven")
		}

		if a["password"] == "" {
			return fmt.Errorf("expected to get a service user password from Aiven")
		}

		if a["type"] == "" {
			return fmt.Errorf("expected to get a service user type from Aiven")
		}

		if a["project"] == "" {
			return fmt.Errorf("expected to get a service user project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service user service_name from Aiven")
		}

		return nil
	}
}
