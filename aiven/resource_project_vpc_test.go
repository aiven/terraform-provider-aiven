package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAivenProjectVPC_basic(t *testing.T) {
	resourceName := "aiven_project_vpc.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenProjectVPCResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectVPCResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectVPCAttributes("data.aiven_project_vpc.vpc"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
				),
			},
			{
				Config: testAccProjectVPCCustomTimeoutResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenProjectVPCAttributes("data.aiven_project_vpc.vpc"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "network_cidr", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
				),
			},
		},
	})
}

func testAccProjectVPCResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		resource "aiven_project_vpc" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			network_cidr = "192.168.0.0/24"
		}

		data "aiven_project_vpc" "vpc" {
			project = aiven_project_vpc.bar.project
			cloud_name = aiven_project_vpc.bar.cloud_name
		}
		`, name)
}

func testAccProjectVPCCustomTimeoutResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		resource "aiven_project_vpc" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			network_cidr = "192.168.0.0/24"
		}

		data "aiven_project_vpc" "vpc" {
			project = aiven_project_vpc.bar.project
			cloud_name = aiven_project_vpc.bar.cloud_name
			client_create_wait_timeout = 80
		}
		`, name)
}

func testAccCheckAivenProjectVPCAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["cloud_name"] == "" {
			return fmt.Errorf("expected to get an project user cloud_name from Aiven")
		}

		if a["network_cidr"] == "" {
			return fmt.Errorf("expected to get an project user network_cidr from Aiven")
		}

		if a["state"] == "" {
			return fmt.Errorf("expected to get an project user state from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenProjectVPCResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each project VPC is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_project_vpc" {
			continue
		}

		projectName, vpcId := splitResourceID2(rs.Primary.ID)
		vpc, err := c.VPCs.Get(projectName, vpcId)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if vpc != nil {
			return fmt.Errorf("porject vpc (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func Test_copyVPCPropertiesFromAPIResponseToTerraform(t *testing.T) {
	type args struct {
		d       *schema.ResourceData
		vpc     *aiven.VPC
		project string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"basic",
			args{
				d: resourceProjectVPC().Data(&terraform.InstanceState{
					ID:         "",
					Attributes: nil,
					Ephemeral:  terraform.EphemeralState{},
					Meta:       nil,
					Tainted:    false,
				}),
				vpc: &aiven.VPC{
					CloudName:    "google-europe-west1",
					ProjectVPCID: "123",
					State:        "APPROVE",
				},
				project: "test-pr1",
			},
			false,
		},
		{
			"missing-project",
			args{
				d: testProjectVPCResourceMissingField("project"),
				vpc: &aiven.VPC{
					CloudName:    "google-europe-west1",
					ProjectVPCID: "123",
					State:        "APPROVE",
				},
				project: "test-pr1",
			},
			true,
		},
		{
			"missing-cloud_name",
			args{
				d: testProjectVPCResourceMissingField("cloud_name"),
				vpc: &aiven.VPC{
					CloudName:    "google-europe-west1",
					ProjectVPCID: "123",
					State:        "APPROVE",
				},
				project: "test-pr1",
			},
			true,
		},
		{
			"missing-state",
			args{
				d: testProjectVPCResourceMissingField("state"),
				vpc: &aiven.VPC{
					CloudName:    "google-europe-west1",
					ProjectVPCID: "123",
					State:        "APPROVE",
				},
				project: "test-pr1",
			},
			true,
		},
		{
			"missing-network_cidr",
			args{
				d: testProjectVPCResourceMissingField("network_cidr"),
				vpc: &aiven.VPC{
					CloudName:    "google-europe-west1",
					ProjectVPCID: "123",
					State:        "APPROVE",
				},
				project: "test-pr1",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := copyVPCPropertiesFromAPIResponseToTerraform(tt.args.d, tt.args.vpc, tt.args.project); (err != nil) != tt.wantErr {
				t.Errorf("copyVPCPropertiesFromAPIResponseToTerraform() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testProjectVPCResourceMissingField(missing string) *schema.ResourceData {
	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"project":      aivenProjectVPCSchema["project"],
			"cloud_name":   aivenProjectVPCSchema["cloud_name"],
			"network_cidr": aivenProjectVPCSchema["network_cidr"],
			"state":        aivenProjectVPCSchema["state"],
		},
	}

	delete(res.Schema, missing)

	return res.Data(&terraform.InstanceState{})
}
