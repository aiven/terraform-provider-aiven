// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAivenVPCPeeringConnection_basic(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" ||
		os.Getenv("AWS_VPC_ID") == "" ||
		os.Getenv("AWS_ACCOUNT_ID") == "" {
		t.Skip("env variables AWS_REGION, AWS_VPC_ID and AWS_ACCOUNT_ID required to run this test")
	}

	resourceName := "aiven_vpc_peering_connection.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionAWSResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "peer_cloud_account", os.Getenv("AWS_ACCOUNT_ID")),
					resource.TestCheckResourceAttr(resourceName, "peer_vpc", os.Getenv("AWS_VPC_ID")),
					resource.TestCheckResourceAttr(resourceName, "peer_region", os.Getenv("AWS_REGION")),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionAWSResource() string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_project_vpc" "bar" {
		  project      = data.aiven_project.foo.project
		  cloud_name   = "aws-%s"
		  network_cidr = "10.0.0.0/24"
		
		  timeouts {
		    create = "5m"
		  }
		}
		
		resource "aiven_vpc_peering_connection" "foo" {
		  vpc_id             = aiven_project_vpc.bar.id
		  peer_cloud_account = "%s"
		  peer_vpc           = "%s"
		  peer_region        = "%s"
		
		  timeouts {
		    create = "10m"
		  }
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"),
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_ACCOUNT_ID"),
		os.Getenv("AWS_VPC_ID"),
		os.Getenv("AWS_REGION"))
}

func Test_convertStateInfoToMap(t *testing.T) {
	type args struct {
		s *map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "basic",
			args: args{
				&map[string]interface{}{
					"message": "xxx",
					"type":    "xxx",
					"warnings": []interface{}{
						map[string]interface{}{
							"field_a": "xxx",
							"message": "xxx",
							"type":    "overlapping-peer-vpc-ip-ranges"},
					},
				},
			},
			want: map[string]string{
				"message":  "xxx",
				"type":     "xxx",
				"warnings": "[map[field_a:xxx message:xxx type:overlapping-peer-vpc-ip-ranges]]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertStateInfoToMap(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertStateInfoToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
