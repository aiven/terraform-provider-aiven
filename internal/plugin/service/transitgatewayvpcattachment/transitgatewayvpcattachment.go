package transitgatewayvpcattachment

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.ModifyPlan = modifyPlan
	// CIDR updates persist synchronously; the API does not expose cloud route readiness.
	ResourceOptions.RefreshStateCheck = pluginvpc.PeeringCreateStateCheck
	DataSourceOptions.Read = datasourceReadView
}

// parsePeeringID preserves whether the region was explicitly supplied: the SDK
// also populated peer_region from the API when the four-part ID omitted it.
func parsePeeringID(value string) (*pluginvpc.ProjectPeeringID, error) {
	id, err := pluginvpc.ParseProjectPeeringID(value)
	if err != nil {
		return nil, err
	}
	// TGW rejects an empty fifth component, legacy cloud-specific resources accept it.
	if id.PeerRegion != nil && *id.PeerRegion == "" {
		return nil, errors.New("peer region in the fifth ID component must not be empty")
	}
	return id, nil
}

func expandPeeringID(d adapter.ResourceData) (*pluginvpc.ProjectPeeringID, error) {
	id, err := pluginvpc.NewProjectPeeringID(d, "peer_cloud_account", "peer_vpc")
	if err != nil {
		return nil, err
	}
	if peerRegion, ok := d.GetOk("peer_region"); ok && peerRegion.(string) != "" {
		id.PeerRegion = new(peerRegion.(string))
	}
	return id, nil
}

func createView(ctx context.Context, cl avngen.Client, d adapter.ResourceData) error {
	id, err := expandPeeringID(d)
	if err != nil {
		return err
	}

	cidrs := lo.Must(lo.FromAnySlice[string](d.Get("user_peer_network_cidrs").([]any)))
	var userPeerNetworkCidrs *[]string
	if len(cidrs) > 0 {
		userPeerNetworkCidrs = &cidrs
	}

	if err = id.Create(ctx, cl, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount:     id.PeerCloudAccount,
		PeerRegion:           id.PeerRegion,
		PeerVpc:              id.PeerVPC,
		UserPeerNetworkCidrs: userPeerNetworkCidrs,
	}, true); err != nil {
		return fmt.Errorf("creating VPC peering connection: %w", err)
	}
	return d.SetID(id.String())
}

func readView(ctx context.Context, cl avngen.Client, d adapter.ResourceData) error {
	id, err := parsePeeringID(d.ID())
	if err != nil {
		return fmt.Errorf("parsing peering VPC ID: %w", err)
	}

	conn, err := id.Find(ctx, cl)
	if err != nil {
		return fmt.Errorf("find VPC peering connection by Terraform ID: %w", err)
	}
	if err := pluginvpc.PendingPeerWarning(ctx, "")(d, map[string]any{"state": conn.State, "state_info": conn.StateInfo}); err != nil {
		return err
	}

	configured, ok := d.GetConfigOk("user_peer_network_cidrs")
	if !ok {
		return setConnectionState(d, id, conn)
	}
	desired := lo.Must(lo.FromAnySlice[string](configured.([]any)))

	// CIDRs can be reconciled only for a live AWS Transit Gateway attachment and
	// only while the API value differs from the explicitly configured set.
	if conn.VpcPeeringConnectionType != vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment ||
		!isLiveState(string(conn.State)) ||
		lo.ElementsMatch(conn.UserPeerNetworkCidrs, desired) {
		return setConnectionState(d, id, conn)
	}

	if err := updateCIDRs(ctx, cl, id, conn, desired); err != nil {
		return fmt.Errorf("reconcile configured CIDRs after reading %q: %w", d.ID(), err)
	}
	// PUT persists CIDRs synchronously; cloud routes are applied asynchronously.
	conn.UserPeerNetworkCidrs = desired
	return setConnectionState(d, id, conn)
}

func datasourceReadView(ctx context.Context, cl avngen.Client, d adapter.ResourceData) error {
	id, err := expandPeeringID(d)
	if err != nil {
		return err
	}

	conn, err := id.Find(ctx, cl)
	if err != nil {
		return fmt.Errorf("lookup `aiven_transit_gateway_vpc_attachment` by `peer_cloud_account`, `peer_vpc` and optional `peer_region`: %w", err)
	}

	if conn.PeerRegion != nil && *conn.PeerRegion != "" {
		id.PeerRegion = conn.PeerRegion
	}
	return setConnectionState(d, id, conn)
}

func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	if err := pluginvpc.PlanProjectVPC(d); err != nil {
		return err
	}
	if peer, known := d.GetConfigOk("peer_vpc"); known && strings.HasPrefix(peer.(string), "vpc-") {
		if cidrs, known := d.GetConfigOk("user_peer_network_cidrs"); known && len(cidrs.([]any)) > 0 {
			return errors.New("user_peer_network_cidrs must be empty for ordinary AWS VPC peering connections; configure CIDRs only for a Transit Gateway (tgw-*)")
		}
	}
	// Outside the live states the API intentionally projects CIDRs as an empty set,
	// regardless of whether the underlying rows still exist. Keep state aligned with
	// that projection and suppress the resulting configuration diff: Update rejects
	// terminal connections and cannot make the configured value effective.
	//
	// Terraform re-plans replacements with null prior state, so IsNewResource keeps
	// the configured CIDRs in the replacement plan. Unknown values must also remain
	// unknown rather than being replaced with prior state.
	if d.IsNewResource() || isLiveState(d.GetState("state").(string)) || !d.HasChange("user_peer_network_cidrs") {
		return nil
	}
	if _, known := d.GetOk("user_peer_network_cidrs"); !known {
		return nil
	}
	return d.Set("user_peer_network_cidrs", d.GetState("user_peer_network_cidrs"))
}

func updateView(ctx context.Context, cl avngen.Client, d adapter.ResourceData) error {
	// Only explicitly configured CIDRs can be written back to the API.
	if _, configured := d.GetConfigOk("user_peer_network_cidrs"); !configured {
		return nil
	}
	id, err := parsePeeringID(d.ID())
	if err != nil {
		return fmt.Errorf("error parsing peering VPC ID: %w", err)
	}

	conn, err := id.Find(ctx, cl)
	if err != nil {
		return fmt.Errorf("cannot get transit gateway vpc attachment by id %s: %w", d.ID(), err)
	}

	desired := lo.Must(lo.FromAnySlice[string](d.Get("user_peer_network_cidrs").([]any)))
	if lo.ElementsMatch(conn.UserPeerNetworkCidrs, desired) {
		return nil
	}
	if conn.VpcPeeringConnectionType != vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment {
		return fmt.Errorf("cannot update user_peer_network_cidrs for VPC peering connection type %q", conn.VpcPeeringConnectionType)
	}
	if !isLiveState(string(conn.State)) {
		return fmt.Errorf("cannot update user_peer_network_cidrs while VPC peering connection is in state %q", conn.State)
	}

	return updateCIDRs(ctx, cl, id, conn, desired)
}

func updateCIDRs(ctx context.Context, cl avngen.Client, id *pluginvpc.ProjectPeeringID, conn *vpc.PeeringConnectionOut, desired []string) error {
	add := make([]vpc.AddIn, 0)
	for _, cidr := range desired {
		if !slices.Contains(conn.UserPeerNetworkCidrs, cidr) {
			add = append(add, vpc.AddIn{
				Cidr:             cidr,
				PeerCloudAccount: id.PeerCloudAccount,
				PeerRegion:       conn.PeerRegion,
				PeerVpc:          id.PeerVPC,
			})
		}
	}

	remove := make([]string, 0)
	for _, cidr := range conn.UserPeerNetworkCidrs {
		if !slices.Contains(desired, cidr) {
			remove = append(remove, cidr)
		}
	}

	if len(add) == 0 && len(remove) == 0 {
		return nil
	}

	if _, err := cl.VpcPeeringConnectionUpdate(ctx, id.Project, id.ProjectVPCID, &vpc.VpcPeeringConnectionUpdateIn{Add: &add, Delete: &remove}); err != nil {
		return fmt.Errorf("cannot update transit gateway VPC attachment: %w", err)
	}
	return nil
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := parsePeeringID(d.ID())
	if err != nil {
		return fmt.Errorf("error parsing peering VPC ID: %w", err)
	}

	if id.PeerRegion == nil {
		region, err := id.ResolveRegion(ctx, client)
		if err != nil {
			return err
		}
		// A regionless DELETE targets the VPC's region, while a four-part ID
		// can identify an imported peering in another region.
		if region != "" {
			id.PeerRegion = &region
		}
	}
	return id.Delete(ctx, client, nil)
}

func setConnectionState(d adapter.ResourceData, id *pluginvpc.ProjectPeeringID, conn *vpc.PeeringConnectionOut) error {
	values := map[string]any{
		"peer_cloud_account":      conn.PeerCloudAccount,
		"peer_vpc":                conn.PeerVpc,
		"peering_connection_id":   nil,
		"state":                   string(conn.State),
		"state_info":              pluginvpc.StateInfoMap(conn.StateInfo),
		"user_peer_network_cidrs": conn.UserPeerNetworkCidrs,
		"project":                 id.Project,
		"project_vpc_id":          id.ProjectVPCID,
		"vpc_id":                  schemautil.BuildResourceID(id.Project, id.ProjectVPCID),
	}
	// Terraform distinguishes an explicitly configured empty string from null. Both
	// mean omission during resource Create and produce a four-part ID, but replacing
	// "" with null would be an inconsistent result. Leave an existing empty value
	// untouched because ResourceData.Set also normalizes a top-level empty string to
	// null. A five-part resource ID remains authoritative, so a data source may resolve a
	// region from the API while retaining its explicitly configured empty selector.
	if peerRegion, ok := d.GetOk("peer_region"); !ok || peerRegion.(string) != "" || (id.PeerRegion != nil && !d.IsDataSource()) {
		values["peer_region"] = nil
		// A four-part resource ID means peer_region was omitted from configuration.
		// Clear any region the SDK discovered so an unchanged configuration does not
		// replace the attachment. Five-part IDs retain their explicit region.
		if id.PeerRegion != nil {
			values["peer_region"] = *id.PeerRegion
		}
	}
	if providerID, ok := conn.StateInfo["aws_vpc_peering_connection_id"].(string); ok && providerID != "" {
		values["peering_connection_id"] = providerID
	}
	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}

	return d.SetID(id.String())
}

func isLiveState(state string) bool {
	return state == string(vpc.VpcPeeringConnectionStateTypeApproved) ||
		state == string(vpc.VpcPeeringConnectionStateTypeActive) ||
		state == string(vpc.VpcPeeringConnectionStateTypePendingPeer)
}
