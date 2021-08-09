// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"vpc_id": {
		Description: "The VPC the peering connection belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_cloud_account": {
		Description: "AWS account ID or GCP project ID of the peered VPC",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_vpc": {
		Description: "AWS VPC ID or GCP VPC network name of the peered VPC",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_region": {
		Description: "AWS region of the peered VPC (if not in the same region as Aiven VPC)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == ""
		},
	},
	"state": {
		Computed:    true,
		Description: "State of the peering connection",
		Type:        schema.TypeString,
	},
	"state_info": {
		Computed:    true,
		Description: "State-specific help or error information",
		Type:        schema.TypeMap,
	},
	"peering_connection_id": {
		Computed:    true,
		Description: "Cloud provider identifier for the peering connection if available",
		Type:        schema.TypeString,
	},
	"peer_azure_app_id": {
		Optional:    true,
		ForceNew:    true,
		Description: "Azure app registration id in UUID4 form that is allowed to create a peering to the peer vnet",
		Type:        schema.TypeString,
	},
	"peer_azure_tenant_id": {
		Optional:    true,
		ForceNew:    true,
		Description: "Azure tenant id in UUID4 form",
		Type:        schema.TypeString,
	},
	"peer_resource_group": {
		Optional:    true,
		ForceNew:    true,
		Description: "Azure resource group name of the peered VPC",
		Type:        schema.TypeString,
	},
}

func resourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPeeringConnectionCreate,
		ReadContext:   resourceVPCPeeringConnectionRead,
		DeleteContext: resourceVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPeeringConnectionImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenVPCPeeringConnectionSchema,
	}
}

func resourceVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		pc     *aiven.VPCPeeringConnection
		err    error
		region *string
		cidrs  []string
	)

	client := m.(*aiven.Client)
	projectName, vpcID := splitResourceID2(d.Get("vpc_id").(string))
	if projectName == "" || vpcID == "" {
		return diag.Errorf("incorrect VPC ID, expected structure <PROJECT_NAME>/<VPC_ID>")
	}

	peerCloudAccount := d.Get("peer_cloud_account").(string)
	peerVPC := d.Get("peer_vpc").(string)
	peerRegion := d.Get("peer_region").(string)
	if peerRegion != "" {
		region = &peerRegion
	}

	if userPeerNetworkCidrs, ok := d.GetOk("user_peer_network_cidrs"); ok {
		cidrs = flattenToString(userPeerNetworkCidrs.([]interface{}))
	}

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

	if _, err = client.VPCPeeringConnections.Create(
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount:     peerCloudAccount,
			PeerVPC:              peerVPC,
			PeerRegion:           region,
			UserPeerNetworkCIDRs: cidrs,
			PeerAzureAppId:       peerAzureAppID,
			PeerAzureTenantId:    peerAzureTenantID,
			PeerResourceGroup:    peerResourceGroup,
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
				peerCloudAccount,
				peerVPC,
				region,
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

	diags := diag.Diagnostics{}
	pc = res.(*aiven.VPCPeeringConnection)
	if pc.State != "ACTIVE" {
		switch pc.State {
		case "PENDING_PEER":
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary: fmt.Sprintf("Aiven platform has created a connection to the specified "+
					"peer successfully in the cloud, but the connection is not active until the user "+
					"completes the setup in their cloud account. The steps needed in the user cloud "+
					"account depend on the used cloud provider. Find more in the state info: %s",
					stateInfoToString(pc.StateInfo)),
			})
		case "DELETED":
			diags = append(diags, diag.Errorf("A user has deleted the peering connection through the Aiven "+
				"Terraform provider, or Aiven Web Console or directly via Aiven API. There are no "+
				"transitions from this state")...)
		case "DELETED_BY_PEER":
			diags = append(diags, diag.Errorf("A user deleted the peering cloud resource in their account. "+
				"There are no transitions from this state")...)
		case "REJECTED_BY_PEER":
			diags = append(diags, diag.Errorf("AWS VPC peering connection request was rejected, state info: %s",
				stateInfoToString(pc.StateInfo))...)
		case "INVALID_SPECIFICATION":
			diags = append(diags, diag.Errorf("VPC peering connection cannot be created, more in the state info: %s",
				stateInfoToString(pc.StateInfo))...)
		default:
			return diag.Errorf("Unknown VPC peering connection cannot state: %s", pc.State)
		}
	}

	if peerRegion != "" {
		d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC, *pc.PeerRegion))
	} else {
		d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))
	}

	// in case of an error delete VPC peering connection
	if diags.HasError() {
		return append(diags, resourceVPCPeeringConnectionDelete(ctx, d, m)...)
	}

	return append(diags, resourceVPCPeeringConnectionRead(ctx, d, m)...)
}

// stateInfoToString converts VPC peering connection state_info to a string
func stateInfoToString(s *map[string]interface{}) string {
	if len(*s) == 0 {
		return ""
	}

	var str string
	// Print message first
	if m, ok := (*s)["message"]; ok {
		str = fmt.Sprintf("%s", m)
		delete(*s, "message")
	}

	for k, v := range *s {
		if _, ok := v.(string); ok {
			str += fmt.Sprintf("\n %q:%q", k, v)
		} else {
			str += fmt.Sprintf("\n %q:`%+v`", k, v)
		}
	}

	return str
}

func parsePeeringVPCId(resourceID string) (string, string, string, string, *string) {
	var peerRegion *string

	parts := strings.Split(resourceID, "/")
	projectName := parts[0]
	vpcID := parts[1]
	peerCloudAccount := parts[2]
	peerVPC := parts[3]
	if len(parts) > 4 {
		peerRegion = new(string)
		*peerRegion = parts[4]
	}

	return projectName, vpcID, peerCloudAccount, peerVPC, peerRegion
}

func resourceVPCPeeringConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var pc *aiven.VPCPeeringConnection
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return diag.Errorf("Error checking if it Azure VPC peering connection: %s", err)
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			pc, err = client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
				projectName, vpcID, peerCloudAccount, peerVPC, peerRegion, peerResourceGroup.(string))
			if err != nil {
				return diag.Errorf("Error getting VPC peering connection resource group: %s", err)
			}
		} else {
			return diag.Errorf("cannot get an Azure VPC peering connection without `peer_resource_group`")
		}

		return append(
			copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID),
			copyAzureVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc)...)
	}

	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)
	if err != nil {
		return diag.Errorf("Error getting VPC peering connection: %s", err)
	}

	return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())

	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return diag.Errorf("Error checking if it Azure VPC peering connection: %s", err)
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			if err = client.VPCPeeringConnections.DeleteVPCPeeringWithResourceGroup(
				projectName,
				vpcID,
				peerCloudAccount,
				peerVPC,
				peerResourceGroup.(string),
				peerRegion,
			); err != nil && !aiven.IsNotFound(err) {
				return diag.Errorf("Error deleting VPC peering connection with resource group: %s", err)
			}
		} else {
			return diag.Errorf("cannot delete an Azure VPC peering connection without `peer_resource_group`")
		}
	}
	if err = client.VPCPeeringConnections.DeleteVPCPeering(
		projectName,
		vpcID,
		peerCloudAccount,
		peerVPC,
		peerRegion,
	); err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting VPC peering connection: %s", err)
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
			var pc *aiven.VPCPeeringConnection
			if isAzure {
				pc, err = client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
					projectName,
					vpcID,
					peerCloudAccount,
					peerVPC,
					peerRegion,
					d.Get("peer_resource_group").(string), // was already checked
				)
			} else {
				pc, err = client.VPCPeeringConnections.GetVPCPeering(
					projectName,
					vpcID,
					peerCloudAccount,
					peerVPC,
					peerRegion,
				)
			}
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
		return diag.Errorf("Error waiting for Aiven VPC Peering Connection to be DELETED: %s", err)
	}
	return nil
}

func resourceVPCPeeringConnectionImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 4 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<vpc_id>/<peer_cloud_account>/<peer_vpc>", d.Id())
	}

	dig := resourceVPCPeeringConnectionRead(ctx, d, m)
	if dig.HasError() {
		return nil, errors.New("cannot get VPC peering connection")
	}

	return []*schema.ResourceData{d}, nil
}

func copyAzureVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := d.Set("peer_azure_app_id", peeringConnection.PeerAzureAppId); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_azure_app_id field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_azure_app_id field for VPC peering connection: %s", err),
		})
	}
	if err := d.Set("peer_azure_tenant_id", peeringConnection.PeerAzureTenantId); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_azure_tenant_id field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_azure_tenant_id field for VPC peering connection: %s", err),
		})
	}
	if err := d.Set("peer_resource_group", peeringConnection.PeerResourceGroup); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_resource_group field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_resource_group field for VPC peering connection: %s", err),
		})
	}

	return diags
}

func copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
	project string,
	vpcID string,
) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := d.Set("vpc_id", buildResourceID(project, vpcID)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set vpc_id field: %s", err),
			Detail:   fmt.Sprintf("Unable to set vpc_id field for VPC peering connection: %s", err),
		})
	}
	if err := d.Set("peer_cloud_account", peeringConnection.PeerCloudAccount); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_cloud_account field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_cloud_account field for VPC peering connection: %s", err),
		})
	}
	if err := d.Set("peer_vpc", peeringConnection.PeerVPC); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_vpc field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_vpc field for VPC peering connection: %s", err),
		})
	}
	if peeringConnection.PeerRegion != nil {
		if err := d.Set("peer_region", peeringConnection.PeerRegion); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set peer_region field: %s", err),
				Detail:   fmt.Sprintf("Unable to set peer_region field for VPC peering connection: %s", err),
			})
		}
	}
	if err := d.Set("state", peeringConnection.State); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set state field: %s", err),
			Detail:   fmt.Sprintf("Unable to set state field for VPC peering connection: %s", err),
		})
	}

	if peeringConnection.StateInfo != nil {
		peeringID, ok := (*peeringConnection.StateInfo)["aws_vpc_peering_connection_id"]
		if ok {
			if err := d.Set("peering_connection_id", peeringID); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Unable to set peering_connection_id field: %s", err),
					Detail:   fmt.Sprintf("Unable to set peering_connection_id field for VPC peering connection: %s", err),
				})
			}
		}
	}

	if err := d.Set("state_info", convertStateInfoToMap(peeringConnection.StateInfo)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set state_info field: %s", err),
			Detail:   fmt.Sprintf("Unable to set state_info field for VPC peering connection: %s", err),
		})
	}

	// user_peer_network_cidrs filed is only available for transit gateway vpc attachment
	if len(peeringConnection.UserPeerNetworkCIDRs) > 0 {
		// convert cidrs from []string to []interface {}
		cidrs := make([]interface{}, len(peeringConnection.UserPeerNetworkCIDRs))
		for i, cidr := range peeringConnection.UserPeerNetworkCIDRs {
			cidrs[i] = cidr
		}

		if err := d.Set("user_peer_network_cidrs", cidrs); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set user_peer_network_cidrs field: %s", err),
				Detail:   fmt.Sprintf("Unable to set user_peer_network_cidrs field for VPC peering connection: %s", err),
			})
		}
	}

	return diags
}

func convertStateInfoToMap(s *map[string]interface{}) map[string]string {
	if s == nil || len(*s) == 0 {
		return nil
	}

	r := make(map[string]string)
	for k, v := range *s {
		if _, ok := v.(string); ok {
			r[k] = v.(string)
		} else {
			r[k] = fmt.Sprintf("%+v", v)
		}
	}

	return r
}

// isAzureVPCPeeringConnection checking if peered VPC is in the Azure cloud
func isAzureVPCPeeringConnection(d *schema.ResourceData, c *aiven.Client) (bool, error) {
	projectName, vpcID, _, _, peerRegion := parsePeeringVPCId(d.Id())

	// If peerRegion is nil the peered VPC is assumed to be in the same region and
	// cloud as the project VPC
	if peerRegion == nil {
		vpc, err := c.VPCs.Get(projectName, vpcID)
		if err != nil {
			return false, err
		}

		// Project VPC CloudName has [cloud]-[region] structure
		if strings.Contains(vpc.CloudName, "azure") {
			return true, nil
		}

		return false, nil
	}

	if strings.Contains(*peerRegion, "azure") {
		return true, nil
	}

	return false, nil
}
