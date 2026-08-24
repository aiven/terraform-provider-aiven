package vpc

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	ctx := context.Background()

	sweep.AddTestSweepers("aiven_aws_org_vpc_peering_connection", &resource.Sweeper{
		Name: "aiven_aws_org_vpc_peering_connection",
		F:    sweepOrgVPCPeeringCons(ctx),
		Dependencies: []string{
			"aiven_organization_vpc",
		},
	})

	sweep.AddTestSweepers("aiven_azure_org_vpc_peering_connection", &resource.Sweeper{
		Name: "aiven_azure_org_vpc_peering_connection",
		F:    sweepOrgVPCPeeringCons(ctx),
		Dependencies: []string{
			"aiven_organization_vpc",
		},
	})

	sweep.AddTestSweepers("aiven_gcp_org_vpc_peering_connection", &resource.Sweeper{
		Name: "aiven_gcp_org_vpc_peering_connection",
		F:    sweepOrgVPCPeeringCons(ctx),
		Dependencies: []string{
			"aiven_organization_vpc",
		},
	})
}

func sweepOrgVPCPeeringCons(ctx context.Context) func(string) error {
	return func(_ string) error {
		orgName := os.Getenv("AIVEN_ORGANIZATION_NAME")
		client, err := sweep.SharedGenClient()
		if err != nil {
			return err
		}

		list, err := client.AccountList(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving a list of organizations : %w", err)
		}

		for _, org := range list {
			if org.AccountName != orgName {
				continue
			}

			VPCs, err := client.OrganizationVpcList(ctx, org.OrganizationId)
			if common.IsCritical(err) {
				return fmt.Errorf("error retrieving a list of vpcs for a project : %w", err)
			}

			for _, vpc := range VPCs {
				orgVPC, err := client.OrganizationVpcGet(ctx, org.OrganizationId, vpc.OrganizationVpcId)
				if common.IsCritical(err) {
					return fmt.Errorf("error retrieving a list of vpcs for a project : %w", err)
				}

				for _, peeringCon := range orgVPC.PeeringConnections {
					if peeringCon.PeeringConnectionId == nil {
						continue // should not happen
					}

					_, err = client.OrganizationVpcPeeringConnectionDeleteById(ctx, org.OrganizationId, vpc.OrganizationVpcId, *peeringCon.PeeringConnectionId)
					if common.IsCritical(err) {
						return fmt.Errorf("error deleting vpc peering connection %s/%s/%s: %w",
							org.OrganizationId,
							vpc.OrganizationVpcId,
							*peeringCon.PeeringConnectionId,
							err)
					}

				}

			}
		}

		return nil
	}
}
