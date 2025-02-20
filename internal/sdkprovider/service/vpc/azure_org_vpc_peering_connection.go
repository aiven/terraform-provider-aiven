package vpc

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAzureOrgVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Identifier of the organization.",
	},
	"organization_vpc_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Identifier of the organization VPC.",
	},
	"azure_subscription_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The ID of the Azure subscription in UUID4 format.").ForceNew().Build(),
	},
	"vnet_name": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The name of the Azure VNet.").ForceNew().Build(),
	},
	"peer_resource_group": {
		Required:    true,
		ForceNew:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The name of the Azure resource group associated with the VNet.").ForceNew().Build(),
	},
	"peer_azure_app_id": {
		Required:    true,
		ForceNew:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The ID of the Azure app that is allowed to create a peering to the Azure Virtual Network (VNet) in UUID4 format.").ForceNew().Build(),
	},
	"peer_azure_tenant_id": {
		Required:    true,
		ForceNew:    true,
		Type:        schema.TypeString,
		Description: userconfig.Desc("The Azure tenant ID in UUID4 format.").ForceNew().Build(),
	},
	"state": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "State of the peering connection",
	},
	"peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "The ID of the cloud provider for the peering connection.",
	},
}

func ResourceAzureOrgVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Azure VPC peering connection with an Aiven VPC.",
		CreateContext: common.WithGenClientDiag(resourceAzureOrgVPCPeeringConnectionCreate),
		ReadContext:   common.WithGenClientDiag(resourceAzureOrgVPCPeeringConnectionRead),
		DeleteContext: common.WithGenClientDiag(resourceAzureOrgVPCPeeringConnectionDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAzureOrgVPCPeeringConnectionSchema,
	}
}

func resourceAzureOrgVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	var (
		orgID               = d.Get("organization_id").(string)
		vpcID               = d.Get("organization_vpc_id").(string)
		azureSubscriptionID = d.Get("azure_subscription_id").(string)
		vnetName            = d.Get("vnet_name").(string)
		appID               = d.Get("peer_azure_app_id").(string)
		tenantID            = d.Get("peer_azure_tenant_id").(string)
		resourceGroup       = d.Get("peer_resource_group").(string)

		req = organizationvpc.OrganizationVpcPeeringConnectionCreateIn{
			PeerAzureAppId:    util.ToPtr(appID),
			PeerAzureTenantId: util.ToPtr(tenantID),
			PeerCloudAccount:  azureSubscriptionID,
			PeerResourceGroup: util.ToPtr(resourceGroup),
			PeerVpc:           vnetName,
		}
	)

	pCon, err := createPeeringConnection(ctx, orgID, vpcID, client, d, req)
	if err != nil {
		return diag.Errorf("Error creating VPC peering connection: %s", err)
	}

	diags := getDiagnosticsFromState(newOrganizationVPCPeeringState(pCon))

	d.SetId(schemautil.BuildResourceID(orgID, vpcID, pCon.PeerCloudAccount, pCon.PeerVpc, pCon.PeerResourceGroup))

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		deleteDiags := resourceAzureOrgVPCPeeringConnectionDelete(ctx, d, client)
		d.SetId("") // Clear the ID after delete

		return append(diags, deleteDiags...)
	}

	return append(diags, resourceAzureOrgVPCPeeringConnectionRead(ctx, d, client)...)
}

func resourceAzureOrgVPCPeeringConnectionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	orgID, vpcID, cloudAccount, vnetName, resourceGroup, err := schemautil.SplitResourceID5(d.Id())
	if err != nil {
		return diag.Errorf("error parsing Azure peering VPC ID: %s", err)
	}

	vpc, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
	if err != nil {
		if avngen.IsNotFound(err) {
			return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
		}

		return diag.Errorf("failed to get VPC with ID %q: %s", vpcID, err)
	}

	pc := lookupAzurePeeringConnection(vpc, cloudAccount, vnetName, resourceGroup)
	if pc == nil {
		d.SetId("") // Clear the ID as the resource is not found

		return diag.FromErr(fmt.Errorf("VPC peering connection not found"))
	}

	if err = d.Set("organization_id", orgID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("organization_vpc_id", vpcID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("azure_subscription_id", pc.PeerCloudAccount); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("vnet_name", pc.PeerVpc); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("peer_azure_app_id", pc.PeerAzureAppId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("peer_azure_tenant_id", pc.PeerAzureTenantId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("peer_resource_group", pc.PeerResourceGroup); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("peering_connection_id", *pc.PeeringConnectionId); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("state", string(pc.State)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAzureOrgVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	orgID, vpcID, cloudAccount, vnetName, resourceGroup, err := schemautil.SplitResourceID5(d.Id())
	if err != nil {
		return diag.Errorf("error parsing Azure peering VPC ID: %s", err)
	}

	vpc, err := client.OrganizationVpcGet(ctx, orgID, vpcID)
	if err != nil {
		if avngen.IsNotFound(err) {
			return nil // consider already deleted
		}

		return diag.Errorf("failed to get VPC with ID %q: %s", vpcID, err)
	}

	if err = deletePeeringConnection(
		ctx,
		orgID,
		vpcID,
		client,
		d,
		lookupAzurePeeringConnection(vpc, cloudAccount, vnetName, resourceGroup),
	); err != nil {
		return diag.Errorf("Error deleting Azure Aiven VPC Peering Connection: %s", err)
	}

	return nil
}

func lookupAzurePeeringConnection(
	vpc *organizationvpc.OrganizationVpcGetOut,
	peerCloudAccount, peerVPC, resourceGroup string,
) *organizationvpc.OrganizationVpcGetPeeringConnectionOut {
	for _, pc := range vpc.PeeringConnections {
		if pc.PeerCloudAccount == peerCloudAccount &&
			pc.PeerVpc == peerVPC &&
			pc.PeerResourceGroup == resourceGroup {

			return &pc
		}
	}

	return nil
}
