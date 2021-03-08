package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

func TestAccAivenAWSPrivatelink_basic(t *testing.T) {
	if os.Getenv("AIVEN_AWS_PRIVATELINK_VPCID") == "" ||
		os.Getenv("AIVEN_AWS_PRIVATELINK_PRINCIPAL") == "" {
		t.Skip("AIVEN_AWS_PRIVATELINK_VPCID and AIVEN_AWS_PRIVATELINK_PRINCIPAL env variables are required to run this test")
	}

	resourceName := "aiven_aws_privatelink.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenAWSPrivatelinkResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPrivatelinkResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAWSPrivatelinkAttributes("data.aiven_aws_privatelink.pr"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_service_name"),
					resource.TestCheckResourceAttrSet(resourceName, "aws_service_id"),
				),
			},
		},
	})
}

func testAccCheckAivenAWSPrivatelinkResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each AWS privatelink is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_aws_privatelink" {
			continue
		}

		pv, err := c.AWSPrivatelink.Get(splitResourceID2(rs.Primary.ID))
		if err != nil && !aiven.IsNotFound(err) && err.(aiven.Error).Status != 500 {
			return fmt.Errorf("error getting a AWS Privatelink: %w", err)
		}

		if pv != nil {
			return fmt.Errorf("AWS privatelink (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSPrivatelinkResource(name string) string {
	var principal = os.Getenv("AIVEN_AWS_PRIVATELINK_PRINCIPAL")
	var vpcID = os.Getenv("AIVEN_AWS_PRIVATELINK_VPCID")

	return fmt.Sprintf(`
		data "aiven_project" "foo" {
			project = "%s"
		}

		resource "aiven_kafka" "bar" {
			project = data.aiven_project.foo.project
			cloud_name = "aws-eu-west-1"
			plan = "business-4"
			service_name = "test-acc-sr-%s"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			project_vpc_id = "%s"
			
			kafka_user_config {
				kafka_version = "2.4"
				kafka {
				  group_max_session_timeout_ms = 70000
				  log_retention_bytes = 1000000000
				}
			}
		}
		
		resource "aiven_aws_privatelink" "foo" {
			project = data.aiven_project.foo.project
			service_name = aiven_kafka.bar.service_name
			principals = ["%s"]
		}
		
		data "aiven_aws_privatelink" "pr" {
			project = data.aiven_project.foo.project
			service_name = aiven_kafka.bar.service_name

			depends_on = [aiven_aws_privatelink.foo]
		}
		`, os.Getenv("AIVEN_PROJECT_NAME"), name, vpcID, principal)
}

func testAccCheckAivenAWSPrivatelinkAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["aws_service_name"] == "" {
			return fmt.Errorf("expected to get aws_service_name from Aiven")
		}

		if a["aws_service_id"] == "" {
			return fmt.Errorf("expected to get aws_service_id from Aiven")
		}

		return nil
	}
}
