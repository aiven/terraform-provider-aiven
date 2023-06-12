package vpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func resourceVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		pc     *aiven.VPCPeeringConnection
		err    error
		region *string
		cidrs  []string
	)

	client := m.(*aiven.Client)
	projectName, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	peerCloudAccount := d.Get("peer_cloud_account").(string)
	peerVPC := d.Get("peer_vpc").(string)
	peerRegion := d.Get("peer_region").(string)
	if peerRegion != "" {
		region = &peerRegion
	}

	if cidrsv, ok := d.GetOk("user_peer_network_cidrs"); ok {
		cidrsva, ok := cidrsv.(*schema.Set)
		if ok {
			cidrs = schemautil.FlattenToString(cidrsva.List())
		}
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
		return diag.Errorf("error waiting for VPC peering connection creation: %s", err)
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

	pc = res.(*aiven.VPCPeeringConnection)
	diags := getDiagnosticsFromState(pc)

	if peerRegion != "" {
		d.SetId(schemautil.BuildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC, *pc.PeerRegion))
	} else {
		d.SetId(schemautil.BuildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))
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

type peeringVPCID struct {
	projectName      string
	vpcID            string
	peerCloudAccount string
	peerVPC          string
	peerRegion       *string
}

// parsePeerVPCID splits string id like "my-project/id/id/my-vpc" + optional "/region"
func parsePeerVPCID(src string) (*peeringVPCID, error) {
	chunks := strings.Split(src, "/")
	length := len(chunks)
	if length < 4 || 5 < length {
		return nil, fmt.Errorf("expected unix path-like string with 4-5 chunks, got %d", length)
	}

	pID := &peeringVPCID{
		projectName:      chunks[0],
		vpcID:            chunks[1],
		peerCloudAccount: chunks[2],
		peerVPC:          chunks[3],
	}

	if len(chunks) == 5 {
		pID.peerRegion = &chunks[4]
	}
	return pID, nil
}

func resourceVPCPeeringConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing peering VPC ID: %s", err)
	}

	client := m.(*aiven.Client)
	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return diag.Errorf("Error checking if it Azure VPC peering connection: %s", err)
	}

	var pc *aiven.VPCPeeringConnection
	if isAzure {
		peerResourceGroup := schemautil.OptionalStringPointer(d, "peer_resource_group")
		if peerResourceGroup != nil {
			pc, err = client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
				p.projectName, p.vpcID, p.peerCloudAccount, p.peerVPC, p.peerRegion, peerResourceGroup)
			if err != nil {
				return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
			}
		} else {
			return diag.Errorf("cannot get an Azure VPC peering connection without `peer_resource_group`")
		}

		return append(
			copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, p.projectName, p.vpcID),
			copyAzureSpecificVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc)...)
	}

	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		p.projectName, p.vpcID, p.peerCloudAccount, p.peerVPC, p.peerRegion)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, p.projectName, p.vpcID)
}

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func resourceVPCPeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing peering VPC ID: %s", err)
	}

	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return diag.Errorf("Error checking if it Azure VPC peering connection: %s", err)
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			if err = client.VPCPeeringConnections.DeleteVPCPeeringWithResourceGroup(
				p.projectName,
				p.vpcID,
				p.peerCloudAccount,
				p.peerVPC,
				peerResourceGroup.(string),
				p.peerRegion,
			); err != nil && !aiven.IsNotFound(err) {
				return diag.Errorf("Error deleting VPC peering connection with resource group: %s", err)
			}
		} else {
			return diag.Errorf("cannot delete an Azure VPC peering connection without `peer_resource_group`")
		}
	}
	if err = client.VPCPeeringConnections.DeleteVPCPeering(
		p.projectName,
		p.vpcID,
		p.peerCloudAccount,
		p.peerVPC,
		p.peerRegion,
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
					p.projectName,
					p.vpcID,
					p.peerCloudAccount,
					p.peerVPC,
					p.peerRegion,
					schemautil.OptionalStringPointer(d, "peer_resource_group"), // was already checked
				)
			} else {
				pc, err = client.VPCPeeringConnections.GetVPCPeering(
					p.projectName,
					p.vpcID,
					p.peerCloudAccount,
					p.peerVPC,
					p.peerRegion,
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

func copyAzureSpecificVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
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

	if err := d.Set("vpc_id", schemautil.BuildResourceID(project, vpcID)); err != nil {
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

	if err := d.Set("state_info", ConvertStateInfoToMap(peeringConnection.StateInfo)); err != nil {
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

		if err := d.Set("user_peer_network_cidrs", schema.NewSet(schema.HashString, cidrs)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set user_peer_network_cidrs field: %s", err),
				Detail:   fmt.Sprintf("Unable to set user_peer_network_cidrs field for VPC peering connection: %s", err),
			})
		}
	}

	return diags
}

func ConvertStateInfoToMap(s *map[string]interface{}) map[string]string {
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
	p, err := parsePeerVPCID(d.Id())
	if err != nil {
		return false, fmt.Errorf("error parsing Azure peering VPC ID: %s", err)
	}

	// If peerRegion is nil the peered VPC is assumed to be in the same region and
	// cloud as the project VPC
	if p.peerRegion == nil {
		vpc, err := c.VPCs.Get(p.projectName, p.vpcID)
		if err != nil {
			return false, err
		}

		// Project VPC CloudName has [cloud]-[region] structure
		if strings.Contains(vpc.CloudName, "azure") {
			return true, nil
		}

		return false, nil
	}

	if strings.Contains(*p.peerRegion, "azure") {
		return true, nil
	}

	return false, nil
}

func getDiagnosticsFromState(pc *aiven.VPCPeeringConnection) diag.Diagnostics {
	if pc.State != "ACTIVE" {
		switch pc.State {
		case "PENDING_PEER":
			return diag.Diagnostics{{
				Severity: diag.Warning,
				Summary: fmt.Sprintf("Aiven platform has created a connection to the specified "+
					"peer successfully in the cloud, but the connection is not active until the user "+
					"completes the setup in their cloud account. The steps needed in the user cloud "+
					"account depend on the used cloud provider. Find more in the state info: %s",
					stateInfoToString(pc.StateInfo))}}
		case "DELETED":
			return diag.Errorf("A user has deleted the peering connection through the Aiven " +
				"Terraform provider, or Aiven Web Console or directly via Aiven API. There are no " +
				"transitions from this state")
		case "DELETED_BY_PEER":
			return diag.Errorf("A user deleted the peering cloud resource in their account. " +
				"There are no transitions from this state")
		case "REJECTED_BY_PEER":
			return diag.Errorf("VPC peering connection request was rejected, state info: %s",
				stateInfoToString(pc.StateInfo))
		case "INVALID_SPECIFICATION":
			return diag.Errorf("VPC peering connection cannot be created, more in the state info: %s",
				stateInfoToString(pc.StateInfo))
		default:
			return diag.Errorf("Unknown VPC peering connection state: %s", pc.State)
		}
	}
	return nil
}

func validateVPCID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if len(strings.Split(v, "/")) != 2 {
		errors = append(errors, fmt.Errorf("invalid %v, expected <project_name>/<vpc_id>", k))
		return warnings, errors
	}

	return warnings, errors
}
