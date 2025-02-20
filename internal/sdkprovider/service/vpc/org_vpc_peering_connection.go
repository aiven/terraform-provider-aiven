package vpc

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

var (
	pollInterval = common.DefaultStateChangeMinTimeout
	pollDelay    = 1 * time.Second
)

func createPeeringConnection(
	ctx context.Context,
	orgID, orgVpcID string,
	client avngen.Client,
	d *schema.ResourceData,
	req organizationvpc.OrganizationVpcPeeringConnectionCreateIn,
) (*organizationvpc.OrganizationVpcGetPeeringConnectionOut, error) {
	pc, err := client.OrganizationVpcPeeringConnectionCreate(
		ctx,
		orgID,
		orgVpcID,
		&req,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating VPC peering connection: %w", err)
	}

	if pc.PeeringConnectionId == nil {
		return nil, fmt.Errorf("error creating VPC peering connection: missing peering connection ID")
	}

	// wait for the VPC peering connection to be approved
	stateChangeConf := &retry.StateChangeConf{
		Target:  []string{""}, // empty target means we don't care about the target state
		Pending: []string{string(organizationvpc.VpcPeeringConnectionStateTypeApproved)},
		Refresh: func() (any, string, error) {
			resp, err := client.OrganizationVpcGet(ctx, orgID, orgVpcID)
			if err != nil {
				return nil, "", fmt.Errorf("error getting VPC: %w", err)
			}

			pCon := lookupPeeringConnection(resp, *pc.PeeringConnectionId)
			if pCon == nil {
				return nil, "", fmt.Errorf("VPC peering connection not found")
			}

			if pCon.State == organizationvpc.VpcPeeringConnectionStateTypeApproved {
				return pCon, string(pCon.State), nil
			}

			return pCon, "", nil // return empty target state to stop the loop
		},
		Delay:      pollDelay,
		MinTimeout: pollInterval,
		Timeout:    d.Timeout(schema.TimeoutCreate),
	}

	resp, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for VPC peering connection to change state: %w", err)
	}

	pCon, ok := resp.(*organizationvpc.OrganizationVpcGetPeeringConnectionOut)
	if !ok {
		return nil, fmt.Errorf("error creating VPC peering connection: invalid response") // this should never happen
	}

	return pCon, nil
}

func deletePeeringConnection(
	ctx context.Context,
	orgID, orgVpcID string,
	client avngen.Client,
	d *schema.ResourceData,
	pc *organizationvpc.OrganizationVpcGetPeeringConnectionOut,
) error {
	if pc == nil {
		return nil // consider already deleted
	}

	_, err := client.OrganizationVpcPeeringConnectionDeleteById(
		ctx,
		orgID,
		orgVpcID,
		*pc.PeeringConnectionId,
	)
	if err != nil {
		if avngen.IsNotFound(err) {
			return nil // consider already deleted
		}

		return fmt.Errorf("error deleting VPC peering connection: %w", err)
	}

	stateChangeConf := &retry.StateChangeConf{
		Target: []string{string(organizationvpc.VpcPeeringConnectionStateTypeDeleted)},
		Refresh: func() (interface{}, string, error) {
			resp, err := client.OrganizationVpcGet(ctx, orgID, orgVpcID)
			if err != nil {
				if avngen.IsNotFound(err) {
					return struct{}{}, string(organizationvpc.VpcPeeringConnectionStateTypeDeleted), nil
				}

				return nil, "", fmt.Errorf("error getting VPC: %w", err)
			}

			pCon := lookupPeeringConnection(resp, *pc.PeeringConnectionId)
			if pCon == nil {
				// return empty struct to signal to the state change function that the resource is deleted
				return struct{}{}, string(organizationvpc.VpcPeeringConnectionStateTypeDeleted), nil
			}

			return pCon, string(pCon.State), nil
		},
		Delay:      pollDelay,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: pollInterval,
	}

	if _, err = stateChangeConf.WaitForStateContext(ctx); err != nil && !avngen.IsNotFound(err) {
		return fmt.Errorf("error waiting for deletion: %w", err)
	}

	return nil
}

func lookupPeeringConnection(
	vpc *organizationvpc.OrganizationVpcGetOut,
	peeringConnectionID string,
) *organizationvpc.OrganizationVpcGetPeeringConnectionOut {
	for _, pCon := range vpc.PeeringConnections {
		if pCon.PeeringConnectionId != nil && *pCon.PeeringConnectionId == peeringConnectionID {
			return &pCon
		}
	}

	return nil
}
