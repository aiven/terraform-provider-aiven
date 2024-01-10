//go:build sweep

package vpc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	client, err := sweep.SharedClient()
	if err != nil {
		panic(fmt.Sprintf("error getting client: %s", err))
	}

	allServices := []string{
		"aiven_pg",
		"aiven_cassandra",
		"aiven_opensearch",
		"aiven_grafana",
		"aiven_influxdb",
		"aiven_redis",
		"aiven_mysql",
		"aiven_kafka",
		"aiven_kafka_connect",
		"aiven_kafka_mirrormaker",
		"aiven_m3db",
		"aiven_m3aggregator",
		"aiven_flink",
		"aiven_clickhouse",
	}

	resource.AddTestSweepers("aiven_project_vpc", &resource.Sweeper{
		Name: "aiven_project_vpc",
		F:    sweepVPCs(ctx, client),
		Dependencies: []string{
			"aiven_project",
		},
	})

	resource.AddTestSweepers("aiven_aws_vpc_peering_connection", &resource.Sweeper{
		Name: "aiven_aws_vpc_peering_connection",
		F:    sweepVPCPeeringCons(ctx, client),
		Dependencies: []string{
			"aiven_project_vpc",
		},
	})

	resource.AddTestSweepers("aiven_azure_vpc_peering_connection", &resource.Sweeper{
		Name: "aiven_azure_vpc_peering_connection",
		F:    sweepVPCPeeringCons(ctx, client),
		Dependencies: []string{
			"aiven_project_vpc",
		},
	})

	resource.AddTestSweepers("aiven_gcp_vpc_peering_connection", &resource.Sweeper{
		Name: "aiven_gcp_vpc_peering_connection",
		F:    sweepVPCPeeringCons(ctx, client),
		Dependencies: []string{
			"aiven_project_vpc",
		},
	})

	resource.AddTestSweepers("aiven_transit_gateway_vpc_attachment", &resource.Sweeper{
		Name: "aiven_transit_gateway_vpc_attachment",
		F:    sweepVPCPeeringCons(ctx, client),
		Dependencies: []string{
			"aiven_project_vpc",
		},
	})

	resource.AddTestSweepers("aiven_aws_privatelink", &resource.Sweeper{
		Name:         "aiven_aws_privatelink",
		F:            sweepAWSPrivatelinks(ctx, client),
		Dependencies: allServices,
	})

	resource.AddTestSweepers("aiven_azure_privatelink", &resource.Sweeper{
		Name:         "aiven_azure_privatelink",
		F:            sweepAzurePrivatelinks(ctx, client),
		Dependencies: allServices,
	})

	resource.AddTestSweepers("aiven_gcp_privatelink", &resource.Sweeper{
		Name:         "aiven_gcp_privatelink",
		F:            sweepGCPPrivatelinks(ctx, client),
		Dependencies: allServices,
	})
}

func sweepVPCs(ctx context.Context, client *aiven.Client) func(string) error {
	return func(region string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		vpcs, err := client.VPCs.List(ctx, projectName)
		if err != nil {
			return fmt.Errorf("error retrieving a list of vpcs for a project : %s", err)
		}

		for _, vpc := range vpcs {

			err := client.VPCs.Delete(ctx, projectName, vpc.ProjectVPCID)
			if common.IsCritical(err) {
				return fmt.Errorf("error deleting vpc %s: %s", vpc.ProjectVPCID, err)
			}
		}

		return nil
	}
}

func sweepVPCPeeringCons(ctx context.Context, client *aiven.Client) func(string) error {
	return func(region string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		vpcs, err := client.VPCs.List(ctx, projectName)
		if err != nil {
			return fmt.Errorf("error retrieving a list of vpcs for a project : %s", err)
		}

		for _, vpc := range vpcs {
			peeringCons, err := client.VPCPeeringConnections.List(ctx, projectName, vpc.ProjectVPCID)
			if err != nil {
				return fmt.Errorf("error retrieving a list of vpc peering connections for a project : %s", err)
			}

			for _, peeringCon := range peeringCons {
				// If it is azure, we might have a transit gateway attachment,
				// then we need to delete the peering connection with resource group.
				if strings.Contains(vpc.CloudName, "azure") {
					err := client.VPCPeeringConnections.DeleteVPCPeeringWithResourceGroup(
						ctx,
						projectName,
						vpc.ProjectVPCID,
						peeringCon.PeerCloudAccount,
						peeringCon.PeerVPC,
						*peeringCon.PeerResourceGroup,
						peeringCon.PeerRegion)
					if common.IsCritical(err) {
						return fmt.Errorf("error deleting vpc peering connection %s/%s/%s/%s: %s",
							vpc.ProjectVPCID,
							peeringCon.PeerCloudAccount,
							peeringCon.PeerVPC,
							peeringCon.PeerResourceGroup,
							err)
					}
				}

				err := client.VPCPeeringConnections.DeleteVPCPeering(
					ctx,
					projectName,
					vpc.ProjectVPCID,
					peeringCon.PeerCloudAccount,
					peeringCon.PeerVPC,
					peeringCon.PeerRegion)
				if common.IsCritical(err) {
					return fmt.Errorf("error deleting vpc peering connection %s/%s/%s/%s: %s",
						vpc.ProjectVPCID,
						peeringCon.PeerCloudAccount,
						peeringCon.PeerVPC,
						peeringCon.PeerRegion,
						err)
				}
			}
		}

		return nil
	}
}

func sweepAWSPrivatelinks(ctx context.Context, client *aiven.Client) func(string) error {
	return func(region string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		serviceList, err := client.Services.List(ctx, projectName)
		if err != nil {
			return fmt.Errorf("error retrieving a list of services for a project : %s", err)
		}

		for _, service := range serviceList {
			awsPrivetelink, err := client.AWSPrivatelink.Get(ctx, projectName, service.Name)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of aws privatelinks for a project : %s", err)
			}

			if awsPrivetelink != nil {
				err := client.AWSPrivatelink.Delete(ctx, projectName, service.Name)
				if common.IsCritical(err) {
					return fmt.Errorf("error deleting aws privatelink %s/%s: %s",
						projectName,
						service.Name,
						err)
				}
			}
		}

		return nil
	}
}

func sweepAzurePrivatelinks(ctx context.Context, client *aiven.Client) func(string) error {
	return func(region string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		serviceList, err := client.Services.List(ctx, projectName)
		if err != nil {
			return fmt.Errorf("error retrieving a list of services for a project : %s", err)
		}

		for _, service := range serviceList {
			azurePrivetelink, err := client.AzurePrivatelink.Get(ctx, projectName, service.Name)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of azure privatelinks for a project : %s", err)
			}

			if azurePrivetelink != nil {
				err := client.AzurePrivatelink.Delete(ctx, projectName, service.Name)
				if common.IsCritical(err) {
					return fmt.Errorf("error deleting azure privatelink %s/%s: %s",
						projectName,
						service.Name,
						err)
				}
			}
		}

		return nil
	}
}

func sweepGCPPrivatelinks(ctx context.Context, client *aiven.Client) func(string) error {
	return func(region string) error {
		projectName := os.Getenv("AIVEN_PROJECT_NAME")
		serviceList, err := client.Services.List(ctx, projectName)
		if err != nil {
			return fmt.Errorf("error retrieving a list of services for a project : %s", err)
		}

		for _, service := range serviceList {
			gcpPrivetelink, err := client.GCPPrivatelink.Get(ctx, projectName, service.Name)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of gcp privatelinks for a project : %s", err)
			}

			if gcpPrivetelink != nil {
				err := client.GCPPrivatelink.Delete(ctx, projectName, service.Name)
				if common.IsCritical(err) {
					return fmt.Errorf("error deleting gcp privatelink %s/%s: %s",
						projectName,
						service.Name,
						err)
				}
			}
		}

		return nil
	}
}