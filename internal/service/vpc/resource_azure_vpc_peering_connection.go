package vpc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenAzureVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"vpc_id": {
		ForceNew:     true,
		Required:     true,
		Type:         schema.TypeString,
		Description:  schemautil.Complex("The VPC the peering connection belongs to.").ForceNew().Build(),
		ValidateFunc: validateVPCID,
	},
	"azure_subscription_id": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("Azure Subscription ID.").ForceNew().Build(),
	},
	"vnet_name": {
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("Azure Network name.").ForceNew().Build(),
	},
	"peer_resource_group": {
		Required:    true,
		ForceNew:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("Azure resource group name of the peered VPC.").ForceNew().Build(),
	},
	"state": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "State of the peering connection",
	},
	"state_info": {
		Computed:    true,
		Type:        schema.TypeMap,
		Description: "State-specific help or error information",
	},
	"peering_connection_id": {
		Computed:    true,
		Type:        schema.TypeString,
		Description: "Cloud provider identifier for the peering connection if available",
	},
	"peer_azure_app_id": {
		Required:    true,
		ForceNew:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("Azure app registration id in UUID4 form that is allowed to create a peering to the peer vnet.").ForceNew().Build(),
	},
	"peer_azure_tenant_id": {
		Required:    true,
		ForceNew:    true,
		Type:        schema.TypeString,
		Description: schemautil.Complex("Azure tenant id in UUID4 form.").ForceNew().Build(),
	},
}

func ResourceAzureVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Description:   "The Azure VPC Peering Connection resource allows the creation and management of Aiven VPC Peering Connections.",
		CreateContext: resourceAzureVPCPeeringConnectionCreate,
		ReadContext:   resourceAzureVPCPeeringConnectionRead,
		DeleteContext: resourceAzureVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema:        aivenAzureVPCPeeringConnectionSchema,
		CustomizeDiff: customdiff.If(schemautil.ResourceShouldNotExist, vpcCustomDiffAzurePeeringConnectionExists()),
	}
}

func vpcCustomDiffAzurePeeringConnectionExists() func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	return func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
		client := m.(*aiven.Client)

		projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
		if err != nil {
			return err
		}

		azureSubscriptionId := d.Get("azure_subscription_id").(string)
		vnetName := d.Get("vnet_name").(string)
		peerResourceGroup := d.Get("peer_resource_group").(string)

		pc, err := client.VPCPeeringConnections.GetVPCPeering(
			projectName, vpcID, azureSubscriptionId, vnetName, &peerResourceGroup)
		if err != nil && !aiven.IsNotFound(err) {
			return err
		}

		if pc != nil {
			return fmt.Errorf("azure peering connection already exists and cannot be created")
		}

		return nil
	}
}

func resourceAzureVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	azureSubscriptionId := d.Get("azure_subscription_id").(string)
	vnetName := d.Get("vnet_name").(string)

	// Azure related fields are only available for VPC Peering Connection resource but
	// not for Transit Gateway VPC Attachment therefore ResourceData.Get retrieves nil
	// for fields that are not present in the schema.
	var peerAzureAppID, peerAzureTenantID, peerResourceGroup string
	if v, ok := d.GetOk("peer_azure_app_id"); ok {
		peerAzureAppID = v.(string)
	}
	if v, ok := d.GetOk("peer_azure_tenant_id"); ok {
		peerAzureTenantID = v.(string)
	}
	if v, ok := d.GetOk("peer_resource_group"); ok {
		peerResourceGroup = v.(string)
	}

	if _, err := client.VPCPeeringConnections.Create(
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount:  azureSubscriptionId,
			PeerVPC:           vnetName,
			PeerAzureAppId:    peerAzureAppID,
			PeerAzureTenantId: peerAzureTenantID,
			PeerResourceGroup: peerResourceGroup,
		},
	); err != nil {
		return diag.Errorf("Error waiting for VPC peering connection creation: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{"APPROVED"},
		Target: []string{
			"ACTIVE",
			"REJECTED_BY_PEER",
			"PENDING_PEER",
			"INVALID_SPECIFICATION",
			"DELETING",
			"DELETED",
			"DELETED_BY_PEER",
		},
		Refresh: func() (interface{}, string, error) {
			pc, err := client.VPCPeeringConnections.GetVPCPeering(
				projectName,
				vpcID,
				azureSubscriptionId,
				vnetName,
				nil,
			)
			if err != nil {
				return nil, "", err
			}
			return pc, pc.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 2 * time.Second,
	}

	res, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error creating VPC peering connection: %s", err)
	}

	pc := res.(*aiven.VPCPeeringConnection)
	diags := getDiagnosticsFromState(pc)

	d.SetId(schemautil.BuildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		return append(diags, resourceAzureVPCPeeringConnectionDelete(ctx, d, m)...)
	}

	return append(diags, resourceAzureVPCPeeringConnectionRead(ctx, d, m)...)
}

func resourceAzureVPCPeeringConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())

	peerResourceGroup := d.Get("peer_resource_group")
	pc, err := client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
		projectName, vpcID, peerCloudAccount, peerVPC, peerRegion, peerResourceGroup.(string))
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	return copyAzureVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceAzureVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())

	peerResourceGroup := d.Get("peer_resource_group")
	err := client.VPCPeeringConnections.DeleteVPCPeeringWithResourceGroup(
		projectName,
		vpcID,
		peerCloudAccount,
		peerVPC,
		peerResourceGroup.(string),
		peerRegion,
	)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting VPC peering connection with resource group: %s", err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{
			"ACTIVE",
			"APPROVED",
			"APPROVED_PEER_REQUESTED",
			"DELETING",
			"INVALID_SPECIFICATION",
			"PENDING_PEER",
			"REJECTED_BY_PEER",
			"DELETED_BY_PEER",
		},
		Target: []string{
			"DELETED",
		},
		Refresh: func() (interface{}, string, error) {
			pc, err := client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
				projectName,
				vpcID,
				peerCloudAccount,
				peerVPC,
				peerRegion,
				d.Get("peer_resource_group").(string), // was already checked
			)
			if err != nil {
				return nil, "", err
			}
			return pc, pc.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}
	if _, err := stateChangeConf.WaitForStateContext(ctx); err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error waiting for Azure Aiven VPC Peering Connection to be DELETED: %s", err)
	}
	return nil
}

func copyAzureVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
	project string,
	vpcID string,
) diag.Diagnostics {
	if err := d.Set("vpc_id", schemautil.BuildResourceID(project, vpcID)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_subscription_id", peeringConnection.PeerCloudAccount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vnet_name", peeringConnection.PeerVPC); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", peeringConnection.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("state_info", ConvertStateInfoToMap(peeringConnection.StateInfo)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("peer_azure_app_id", peeringConnection.PeerAzureAppId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("peer_azure_tenant_id", peeringConnection.PeerAzureTenantId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("peer_resource_group", peeringConnection.PeerResourceGroup); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
