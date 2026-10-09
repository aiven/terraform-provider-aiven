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
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	ResourceOptions.ModifyPlan = modifyPlan
	ResourceOptions.RefreshStateCheck = refreshStateCheck
	DataSourceOptions.Read = datasourceReadView
}

type peeringID struct {
	project          string
	projectVpcID     string
	peerCloudAccount string
	peerVpc          string
	peerRegion       *string
}

// parsePeeringID accepts PROJECT/VPC_ID/PEER_CLOUD_ACCOUNT/PEER_VPC and
// PROJECT/VPC_ID/PEER_CLOUD_ACCOUNT/PEER_VPC/PEER_REGION.
func parsePeeringID(src string) (*peeringID, error) {
	parts := strings.Split(src, "/")
	if len(parts) != 4 && len(parts) != 5 {
		return nil, fmt.Errorf("expected unix path-like string with 4-5 chunks, got %d", len(parts))
	}

	id := &peeringID{
		project:          parts[0],
		projectVpcID:     parts[1],
		peerCloudAccount: parts[2],
		peerVpc:          parts[3],
	}
	if len(parts) == 5 {
		if parts[4] == "" {
			return nil, errors.New("peer region in the fifth ID component must not be empty")
		}
		id.peerRegion = &parts[4]
	}
	return id, nil
}

func expandPeeringID(d adapter.ResourceData) (*peeringID, error) {
	project, projectVpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return nil, err
	}

	id := &peeringID{
		project:          project,
		projectVpcID:     projectVpcID,
		peerCloudAccount: d.Get("peer_cloud_account").(string),
		peerVpc:          d.Get("peer_vpc").(string),
	}
	if peerRegion, ok := d.GetOk("peer_region"); ok && peerRegion.(string) != "" {
		id.peerRegion = new(peerRegion.(string))
	}
	return id, nil
}

func flattenPeeringID(d adapter.ResourceData, id *peeringID) error {
	parts := []string{id.project, id.projectVpcID, id.peerCloudAccount, id.peerVpc}
	if id.peerRegion != nil {
		parts = append(parts, *id.peerRegion)
	}
	return d.SetID(schemautil.BuildResourceID(parts...))
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

	if _, err = cl.VpcPeeringConnectionCreate(ctx, id.project, id.projectVpcID, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount:     id.peerCloudAccount,
		PeerRegion:           id.peerRegion,
		PeerVpc:              id.peerVpc,
		UserPeerNetworkCidrs: userPeerNetworkCidrs,
	}); err != nil {
		return fmt.Errorf("creating VPC peering connection: %w", err)
	}

	return flattenPeeringID(d, id)
}

func readView(ctx context.Context, cl avngen.Client, d adapter.ResourceData) error {
	id, err := parsePeeringID(d.ID())
	if err != nil {
		return fmt.Errorf("parsing peering VPC ID: %w", err)
	}

	conn, err := findPeeringConnection(ctx, cl, id)
	if err != nil {
		return fmt.Errorf("find VPC peering connection by Terraform ID: %w", err)
	}
	if conn.State == vpc.VpcPeeringConnectionStateTypePendingPeer {
		detail := "Aiven created its side of the connection, but the connection isn't active until the setup is completed in the peer cloud account."
		if stateInfo := formatStateInfo(stateInfoMap(conn.StateInfo)); stateInfo != "" {
			detail += " State info: " + stateInfo
		}
		adapter.AddWarning(ctx, "VPC peering connection is pending peer setup", detail)
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
	// Keep the planned value while the backend applies the update. Returning the
	// retry sentinel prevents this intermediate response from becoming final state.
	conn.UserPeerNetworkCidrs = desired
	if err := setConnectionState(d, id, conn); err != nil {
		return err
	}
	return fmt.Errorf("waiting for transit gateway VPC attachment CIDRs to converge: %w", adapter.ErrRefreshStateDesired)
}

func datasourceReadView(ctx context.Context, cl avngen.Client, d adapter.ResourceData) error {
	id, err := expandPeeringID(d)
	if err != nil {
		return err
	}

	conn, err := findPeeringConnection(ctx, cl, id)
	if err != nil {
		return fmt.Errorf("lookup `aiven_transit_gateway_vpc_attachment` by `peer_cloud_account`, `peer_vpc` and optional `peer_region`: %w", err)
	}

	if conn.PeerRegion != nil && *conn.PeerRegion != "" {
		id.peerRegion = conn.PeerRegion
	}
	if err := flattenPeeringID(d, id); err != nil {
		return err
	}
	return setConnectionState(d, id, conn)
}

func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
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
	id, err := parsePeeringID(d.ID())
	if err != nil {
		return fmt.Errorf("error parsing peering VPC ID: %w", err)
	}

	conn, err := findPeeringConnection(ctx, cl, id)
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

func updateCIDRs(ctx context.Context, cl avngen.Client, id *peeringID, conn *vpc.PeeringConnectionOut, desired []string) error {
	add := make([]vpc.AddIn, 0)
	for _, cidr := range desired {
		if !slices.Contains(conn.UserPeerNetworkCidrs, cidr) {
			add = append(add, vpc.AddIn{
				Cidr:             cidr,
				PeerCloudAccount: id.peerCloudAccount,
				PeerRegion:       conn.PeerRegion,
				PeerVpc:          id.peerVpc,
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

	if _, err := cl.VpcPeeringConnectionUpdate(ctx, id.project, id.projectVpcID, &vpc.VpcPeeringConnectionUpdateIn{Add: &add, Delete: &remove}); err != nil {
		return fmt.Errorf("cannot update transit gateway VPC attachment: %w", err)
	}

	return nil
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	id, err := parsePeeringID(d.ID())
	if err != nil {
		return fmt.Errorf("error parsing peering VPC ID: %w", err)
	}

	if id.peerRegion == nil {
		_, err := client.VpcPeeringConnectionDelete(ctx, id.project, id.projectVpcID, id.peerCloudAccount, id.peerVpc)
		return err
	}

	_, err = client.VpcPeeringConnectionWithRegionDelete(ctx, id.project, id.projectVpcID, id.peerCloudAccount, id.peerVpc, *id.peerRegion)
	return err
}

func findPeeringConnection(ctx context.Context, cl avngen.Client, id *peeringID) (*vpc.PeeringConnectionOut, error) {
	rsp, err := cl.VpcGet(ctx, id.project, id.projectVpcID)
	if err != nil {
		return nil, err
	}
	conn, err := adapter.FindOne(rsp.PeeringConnections, func(i int) bool {
		candidate := rsp.PeeringConnections[i]
		return candidate.PeerCloudAccount == id.peerCloudAccount &&
			candidate.PeerVpc == id.peerVpc &&
			(id.peerRegion == nil || equalPeerRegions(candidate.PeerRegion, id.peerRegion))
	})
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func equalPeerRegions(left, right *string) bool {
	leftEmpty := left == nil || *left == ""
	rightEmpty := right == nil || *right == ""
	if leftEmpty || rightEmpty {
		return leftEmpty && rightEmpty
	}
	return *left == *right
}

func setConnectionState(d adapter.ResourceData, id *peeringID, conn *vpc.PeeringConnectionOut) error {
	values := map[string]any{
		"peer_cloud_account":      conn.PeerCloudAccount,
		"peer_vpc":                conn.PeerVpc,
		"peering_connection_id":   nil,
		"state":                   string(conn.State),
		"state_info":              stateInfoMap(conn.StateInfo),
		"user_peer_network_cidrs": conn.UserPeerNetworkCidrs,
		"vpc_id":                  schemautil.BuildResourceID(id.project, id.projectVpcID),
	}
	// Terraform distinguishes an explicitly configured empty string from null. Both
	// mean omission during resource Create and produce a four-part ID, but replacing ""
	// with null would be an inconsistent result. Leave an existing empty value untouched
	// because ResourceData.Set also normalizes a top-level empty string to null. A
	// five-part managed-resource ID remains authoritative, so a data source may resolve one
	// from the API while retaining its explicitly configured empty selector.
	if peerRegion, ok := d.GetOk("peer_region"); !ok || peerRegion.(string) != "" || (id.peerRegion != nil && !d.IsDataSource()) {
		values["peer_region"] = nil
		// A four-part resource ID means peer_region was omitted from configuration.
		// Do not turn the region discovered by the API into ForceNew configuration;
		// five-part resource and data source IDs carry the region explicitly.
		if id.peerRegion != nil {
			values["peer_region"] = *id.peerRegion
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

	return nil
}

func stateInfoMap(info map[string]any) map[string]string {
	if len(info) == 0 {
		return nil
	}

	result := make(map[string]string, len(info))
	for key, value := range info {
		if str, ok := value.(string); ok {
			result[key] = str
		} else {
			result[key] = fmt.Sprintf("%+v", value)
		}
	}
	return result
}

func refreshStateCheck(d adapter.ResourceData) error {
	state := d.Get("state").(string)
	stateInfo := lo.MapValues(d.Get("state_info").(map[string]any), func(value any, _ string) string {
		return value.(string)
	})
	detail := formatStateInfo(stateInfo)
	if detail != "" {
		detail = "; state_info: " + detail
	}

	switch vpc.VpcPeeringConnectionStateType(state) {
	case vpc.VpcPeeringConnectionStateTypeActive, vpc.VpcPeeringConnectionStateTypePendingPeer:
		return nil
	case vpc.VpcPeeringConnectionStateTypeApproved, vpc.VpcPeeringConnectionStateTypeApprovedPeerRequested:
		return fmt.Errorf("VPC peering connection is still in transient state %q%s", state, detail)
	case vpc.VpcPeeringConnectionStateTypeDeleted, vpc.VpcPeeringConnectionStateTypeDeleting:
		return fmt.Errorf("%w: VPC peering connection was deleted and cannot become active%s", adapter.ErrRefreshStateFailed, detail)
	case vpc.VpcPeeringConnectionStateTypeDeletedByPeer:
		return fmt.Errorf("%w: peer cloud resource was deleted%s", adapter.ErrRefreshStateFailed, detail)
	case vpc.VpcPeeringConnectionStateTypeRejectedByPeer:
		return fmt.Errorf("%w: VPC peering connection request was rejected by the peer%s", adapter.ErrRefreshStateFailed, detail)
	case vpc.VpcPeeringConnectionStateTypeInvalidSpecification:
		return fmt.Errorf("%w: VPC peering connection specification is invalid%s", adapter.ErrRefreshStateFailed, detail)
	case vpc.VpcPeeringConnectionStateTypeError:
		return fmt.Errorf("%w: VPC peering connection reached ERROR%s", adapter.ErrRefreshStateFailed, detail)
	default:
		// A newly introduced backend state isn't proof that the connection is terminal.
		// Keep polling and preserve the checkpointed resource.
		return fmt.Errorf("unknown VPC peering connection state %q%s", state, detail)
	}
}

func formatStateInfo(info map[string]string) string {
	parts := lo.MapToSlice(info, func(key, value string) string {
		return fmt.Sprintf("%s=%q", key, value)
	})
	slices.Sort(parts)
	return strings.Join(parts, ", ")
}

func isLiveState(state string) bool {
	return state == string(vpc.VpcPeeringConnectionStateTypeApproved) ||
		state == string(vpc.VpcPeeringConnectionStateTypeActive) ||
		state == string(vpc.VpcPeeringConnectionStateTypePendingPeer)
}
