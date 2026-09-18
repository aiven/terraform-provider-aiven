package vpc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// ProjectPeeringID preserves the SDK's project VPC peering identifier, including
// the optional fifth region component accepted by legacy imports.
type ProjectPeeringID struct {
	Project          string
	ProjectVPCID     string
	PeerCloudAccount string
	PeerVPC          string
	PeerRegion       *string
}

func ParseProjectPeeringID(value string) (*ProjectPeeringID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 4 && len(parts) != 5 {
		return nil, fmt.Errorf("expected unix path-like string with 4-5 chunks, got %d", len(parts))
	}
	id := &ProjectPeeringID{
		Project: parts[0], ProjectVPCID: parts[1],
		PeerCloudAccount: parts[2], PeerVPC: parts[3],
	}
	if len(parts) == 5 {
		id.PeerRegion = &parts[4]
	}
	return id, nil
}

func (id *ProjectPeeringID) String() string {
	parts := []string{id.Project, id.ProjectVPCID, id.PeerCloudAccount, id.PeerVPC}
	if id.PeerRegion != nil {
		parts = append(parts, *id.PeerRegion)
	}
	return schemautil.BuildResourceID(parts...)
}

// Find requires a unique natural-key match. A missing region is a wildcard.
// An explicitly empty fifth component matches a regionless connection.
func (id *ProjectPeeringID) Find(ctx context.Context, client avngen.Client) (*vpc.PeeringConnectionOut, error) {
	return id.FindMatching(ctx, client, nil)
}

// FindMatching applies an additional filter before checking for a unique match.
func (id *ProjectPeeringID) FindMatching(ctx context.Context, client avngen.Client, match func(*vpc.PeeringConnectionOut) bool) (*vpc.PeeringConnectionOut, error) {
	rsp, err := client.VpcGet(ctx, id.Project, id.ProjectVPCID)
	if err != nil {
		return nil, err
	}
	connection, err := adapter.FindOne(rsp.PeeringConnections, func(i int) bool {
		connection := &rsp.PeeringConnections[i]
		return connection.PeerCloudAccount == id.PeerCloudAccount && connection.PeerVpc == id.PeerVPC &&
			(id.PeerRegion == nil || equalPeerRegions(connection.PeerRegion, id.PeerRegion)) &&
			(match == nil || match(connection))
	})
	if err != nil {
		return nil, fmt.Errorf("finding VPC peering connection %q: %w", id, err)
	}
	return &connection, nil
}

func equalPeerRegions(left, right *string) bool {
	if left == nil || *left == "" {
		return right == nil || *right == ""
	}
	return right != nil && *left == *right
}

// Create uses the API's idempotent POST, which can also re-request a failed
// connection. Its response is incomplete (in particular, it omits peer_region).
// Callers checkpoint the requested ID and obtain state through VpcGet.
func (id *ProjectPeeringID) Create(ctx context.Context, client avngen.Client, request *vpc.VpcPeeringConnectionCreateIn) error {
	_, err := client.VpcPeeringConnectionCreate(ctx, id.Project, id.ProjectVPCID, request)
	return err
}

func (id *ProjectPeeringID) Delete(ctx context.Context, client avngen.Client, resourceGroup *string) error {
	if resourceGroup != nil {
		_, err := client.VpcPeeringConnectionWithResourceGroupDelete(ctx, id.Project, id.ProjectVPCID, id.PeerCloudAccount, *resourceGroup, id.PeerVPC)
		return err
	}
	if id.PeerRegion != nil {
		_, err := client.VpcPeeringConnectionWithRegionDelete(ctx, id.Project, id.ProjectVPCID, id.PeerCloudAccount, id.PeerVPC, *id.PeerRegion)
		return err
	}
	_, err := client.VpcPeeringConnectionDelete(ctx, id.Project, id.ProjectVPCID, id.PeerCloudAccount, id.PeerVPC)
	return err
}

// ProjectPeeringIDFromConfig expands the public vpc_id (PROJECT/VPC_ID) into API
// path components. The cloud-specific resource adds its region when applicable.
func ProjectPeeringIDFromConfig(d adapter.ResourceData, accountField, networkField string) (*ProjectPeeringID, error) {
	project, vpcID, err := schemautil.SplitResourceID2(d.Get("vpc_id").(string))
	if err != nil {
		return nil, err
	}
	return &ProjectPeeringID{
		Project: project, ProjectVPCID: vpcID,
		PeerCloudAccount: d.Get(accountField).(string), PeerVPC: d.Get(networkField).(string),
	}, nil
}

// ProjectPeeringFields restores the compound VPC ID and legacy state_info format.
func ProjectPeeringFields(ctx context.Context, id *ProjectPeeringID, cloud string) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		PendingPeerWarning(ctx, cloud),
		func(_ adapter.ResourceData, dto map[string]any) error {
			dto["vpc_id"] = schemautil.BuildResourceID(id.Project, id.ProjectVPCID)
			info, _ := dto["state_info"].(map[string]any)
			dto["state_info"] = StateInfoMap(info)
			return nil
		},
	)
}

// StateInfoMap preserves the SDK's string representation of arbitrary API values,
// including null entries, which Terraform's normal string conversion would drop.
func StateInfoMap(info map[string]any) map[string]string {
	if len(info) == 0 {
		return nil
	}
	result := make(map[string]string, len(info))
	for key, value := range info {
		result[key] = fmt.Sprintf("%+v", value)
	}
	return result
}

// FormatStateInfo includes all cloud details in a stable order for diagnostics.
func FormatStateInfo(info map[string]string) string {
	parts := make([]string, 0, len(info))
	for key, value := range info {
		parts = append(parts, fmt.Sprintf("%s=%q", key, value))
	}
	slices.Sort(parts)
	return strings.Join(parts, ", ")
}

// PeeringRefreshStateCheck keeps cloud error details in creation diagnostics.
// It runs during the adapter's post-operation refresh, never during plain Read.
func PeeringRefreshStateCheck(d adapter.ResourceData) error {
	state := d.Get("state").(string)
	detail := FormatStateInfo(StateInfoMap(d.Get("state_info").(map[string]any)))
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
