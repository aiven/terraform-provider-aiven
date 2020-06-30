package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccAivenVPCPeeringConnection_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config:             testAccVPCPeeringConnectionResource(rName),
				Check:              resource.ComposeTestCheckFunc(),
			},
			{
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Config:             testAccVPCPeeringConnectionCustomTimeoutResource(rName),
				Check:              resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccVPCPeeringConnectionResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		resource "aiven_project_vpc" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			network_cidr = "192.168.0.0/24"
		}

		resource "aiven_vpc_peering_connection" "foo" {
			vpc_id = aiven_project_vpc.bar.id
			peer_cloud_account = "<PEER_ACCOUNT_ID>"
			peer_vpc = "google-project1"
			peer_region = "google-europe-west1"
		}
		`, name)
}

func testAccVPCPeeringConnectionCustomTimeoutResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
		}

		resource "aiven_project_vpc" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			network_cidr = "192.168.0.0/24"
		}

		resource "aiven_vpc_peering_connection" "foo" {
			vpc_id = aiven_project_vpc.bar.id
			peer_cloud_account = "<PEER_ACCOUNT_ID>"
			peer_vpc = "google-project1"
			peer_region = "google-europe-west1"
			
			timeouts {
				create = "5m"
			}
		}
		`, name)
}

func Test_copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(t *testing.T) {
	peerRegion := "123"

	type args struct {
		d                 *schema.ResourceData
		peeringConnection *aiven.VPCPeeringConnection
		project           string
		vpcID             string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"basic",
			args{
				d: resourceVPCPeeringConnection().Data(&terraform.InstanceState{}),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			false,
		},
		{
			"missing-vpc_id",
			args{
				d: testVPCPeeringConnectionResourceMissingField("vpc_id"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
		{
			"missing-peer_cloud_account",
			args{
				d: testVPCPeeringConnectionResourceMissingField("peer_cloud_account"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
		{
			"missing-peer_vpc",
			args{
				d: testVPCPeeringConnectionResourceMissingField("peer_vpc"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
		{
			"missing-peer_region",
			args{
				d: testVPCPeeringConnectionResourceMissingField("peer_region"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
					PeerRegion:       &peerRegion,
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
		{
			"missing-state",
			args{
				d: testVPCPeeringConnectionResourceMissingField("state"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
		{
			"missing-state_info",
			args{
				d: testVPCPeeringConnectionResourceMissingField("state_info"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
		{
			"missing-peering_connection_id",
			args{
				d: testVPCPeeringConnectionResourceMissingField("peering_connection_id"),
				peeringConnection: &aiven.VPCPeeringConnection{
					PeerCloudAccount: "google1",
					PeerVPC:          "vpc123",
					State:            "ACTIVE",
					StateInfo: &map[string]interface{}{
						"aws_vpc_peering_connection_id": 123,
					},
				},
				project: "test-pr1",
				vpcID:   "test-vpc1",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(tt.args.d, tt.args.peeringConnection, tt.args.project, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testVPCPeeringConnectionResourceMissingField(missing string) *schema.ResourceData {
	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"vpc_id":                aivenVPCPeeringConnectionSchema["vpc_id"],
			"peer_cloud_account":    aivenVPCPeeringConnectionSchema["peer_cloud_account"],
			"peer_vpc":              aivenVPCPeeringConnectionSchema["peer_vpc"],
			"peer_region":           aivenVPCPeeringConnectionSchema["peer_region"],
			"state":                 aivenVPCPeeringConnectionSchema["state"],
			"state_info":            aivenVPCPeeringConnectionSchema["state_info"],
			"peering_connection_id": aivenVPCPeeringConnectionSchema["peering_connection_id"],
		},
	}

	delete(res.Schema, missing)

	return res.Data(&terraform.InstanceState{})
}
