package vpc_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aiven/aiven-go-client"
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kelseyhightower/envconfig"
)

type gcpSecrets struct {
	AivenProject string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	GCPProjectID string `envconfig:"GCP_PROJECT_ID" required:"true"`
	GCPRegion    string `envconfig:"GCP_REGION" required:"true"`
}

func TestAccAivenGCPPeeringConnection_basic(t *testing.T) {
	var s gcpSecrets

	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values have been provided to establish a GCP VPC peering connection: %s", err)
	}

	importResourceName := "aiven_gcp_vpc_peering_connection.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"google": {
				Source:            "hashicorp/google",
				VersionConstraint: ">=4.0.0,<5.0.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccGCPVPCPeeringConnection(&s),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(importResourceName, "state"),
					resource.TestCheckResourceAttrSet(importResourceName, "self_link"),
					resource.TestCheckResourceAttr("google_compute_network_peering.foo", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.aiven_gcp_vpc_peering_connection.bar", "state", "PENDING_PEER"),
					resource.TestCheckResourceAttrSet("data.aiven_gcp_vpc_peering_connection.bar", "self_link"),
				),
			},
			{
				Config: testAccGCPVPCPeeringConnection(&s),
				Check: func(state *terraform.State) error {
					c := acc.TestAccProvider.Meta().(*aiven.Client)
					p := s.AivenProject

				QueryVpc:
					vpcs, err := c.VPCs.List(p)
					if err != nil {
						return err
					}

					var v *aiven.VPC
					for _, vpc := range vpcs {
						if vpc.CloudName == fmt.Sprintf("google-%s", s.GCPRegion) {
							v = vpc
						}
					}

					if v == nil {
						return errors.New("error getting GCP peering connection, project VPC is empty")
					}

					cons, err := c.VPCPeeringConnections.List(p, v.ProjectVPCID)
					if err != nil {
						return fmt.Errorf("error getting list of peering connections: %w", err)
					}

					if len(cons) == 0 {
						return fmt.Errorf("error getting GCP peering connection, list of peering connections is empty for Project VPC ID (%s)", v.ProjectVPCID)
					}

					for _, pvpc := range cons {
						if pvpc.State != "ACTIVE" {
							t.Logf("GCP VPC peering connection in VPC (%s) is in state %s, waiting for it to be ACTIVE...", v.ProjectVPCID, pvpc.State)
							time.Sleep(10 * time.Second)
							goto QueryVpc
						}
					}

					t.Logf("Hooray! GCP VPC peering connection is ACTIVE!")

					return nil
				},
			},
			{
				ResourceName: importResourceName,
				ImportState:  true,
				Config:       testAccGCPVPCPeeringConnection(&s),
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[importResourceName]
					if !ok {
						return "", fmt.Errorf("expected resource '%s' to be present in the state", importResourceName)
					}
					if _, ok := rs.Primary.Attributes["vpc_id"]; !ok {
						return "", fmt.Errorf("expected resource '%s' to have 'vpc_id' attribute", importResourceName)
					}
					if _, ok := rs.Primary.Attributes["gcp_project_id"]; !ok {
						return "", fmt.Errorf("expected resource '%s' to have 'gcp_project_id' attribute", importResourceName)
					}
					if _, ok := rs.Primary.Attributes["peer_vpc"]; !ok {
						return "", fmt.Errorf("expected resource '%s' to have 'peer_vpc' attribute", importResourceName)
					}
					return rs.Primary.ID, nil
				},
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						for k, i := range s {
							t.Logf("GCP VPC imported (%d): %#+v", k, i.Attributes)
						}
						return fmt.Errorf("expected only one instance to be imported, instead %d were imported", len(s))
					}

					attributes := s[0].Attributes
					vpcId, ok := attributes["vpc_id"]
					if !ok {
						return errors.New("expected 'vpc_id' field to be set")
					}

					gcpProjectID, ok := attributes["gcp_project_id"]
					if !ok {
						return errors.New("expected 'gcp_project_id' field to be set")
					}

					peerVPC, ok := attributes["peer_vpc"]
					if !ok {
						return errors.New("expected 'gcp_project_id' field to be set")
					}

					expectedId := fmt.Sprintf("%s/%s/%s", vpcId, gcpProjectID, peerVPC)
					if !strings.EqualFold(s[0].ID, expectedId) {
						return fmt.Errorf("expected ID to match '%s', but got: %s", expectedId, s[0].ID)
					}

					if _, ok := attributes["self_link"]; !ok {
						return errors.New("expected 'self_link' field to be set")
					}

					if s, ok := attributes["state"]; !ok || s != "ACTIVE" {
						return fmt.Errorf("expected 'state' field to be set and equal to `ACTIVE`, got `%s`", s)
					}

					return nil
				},
			},
		},
	})
}

func testAccGCPVPCPeeringConnection(s *gcpSecrets) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = "%[1]s"
}

provider "google" {
  project = "%[2]s"
  region  = "%[3]s"
}

resource "aiven_project_vpc" "project_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "google-%[3]s"
  network_cidr = "10.0.0.0/24"
}

resource "aiven_gcp_vpc_peering_connection" "foo" {
  vpc_id         = aiven_project_vpc.project_vpc.id
  gcp_project_id = "%[2]s"
  peer_vpc       = "default"
}

data "google_compute_network" "foo" {
  project = "%[2]s"
  name    = "default"
}

resource "google_compute_network_peering" "foo" {
  name         = "acc-test-vpc-peering"
  network      = data.google_compute_network.foo.id
  peer_network = aiven_gcp_vpc_peering_connection.foo.self_link
}

data "aiven_gcp_vpc_peering_connection" "bar" {
  vpc_id         = aiven_gcp_vpc_peering_connection.foo.vpc_id
  gcp_project_id = aiven_gcp_vpc_peering_connection.foo.gcp_project_id
  peer_vpc       = aiven_gcp_vpc_peering_connection.foo.peer_vpc

  depends_on = [google_compute_network_peering.foo]
}
`, s.AivenProject, s.GCPProjectID, s.GCPRegion)
}
