package vpc

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

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

// ResolveRegion returns the ID's region, looking up the connection only when it is nil.
// It does not modify the ID; an explicitly empty region remains authoritative.
func (id *ProjectPeeringID) ResolveRegion(ctx context.Context, client avngen.Client) (string, error) {
	if id.PeerRegion != nil {
		return *id.PeerRegion, nil
	}
	connection, err := id.Find(ctx, client)
	if err != nil {
		return "", err
	}
	if connection.PeerRegion == nil {
		return "", nil
	}
	return *connection.PeerRegion, nil
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

// ErrAmbiguousPeeringRegion reports that the caller must select a region explicitly.
var ErrAmbiguousPeeringRegion = errors.New("ambiguous VPC peering connection region")

// Create checks for existing connections before the API's idempotent POST.
// Adopting one during create-before-destroy would let deletion of the old
// Terraform instance remove the replacement's connection too.
// Callers that allow the API to default the region must set requireUniqueRegion to prevent
// adopting a connection in the region selected by the API.
func (id *ProjectPeeringID) Create(ctx context.Context, client avngen.Client, request *vpc.VpcPeeringConnectionCreateIn, requireUniqueRegion bool) error {
	network, err := client.VpcGet(ctx, id.Project, id.ProjectVPCID)
	if err != nil {
		return fmt.Errorf("checking for an existing VPC peering connection: %w", err)
	}

	resourceGroup := ""
	if request.PeerResourceGroup != nil {
		resourceGroup = *request.PeerResourceGroup
	}
	ambiguousRegion := false
	for _, connection := range network.PeeringConnections {
		if connection.PeerCloudAccount == request.PeerCloudAccount && connection.PeerVpc == request.PeerVpc &&
			connection.PeerResourceGroup == resourceGroup {
			if equalPeerRegions(connection.PeerRegion, request.PeerRegion) {
				return fmt.Errorf("VPC peering connection %q already exists in state %s. Import it to manage the existing connection, or use destroy-before-create when replacing it", id, connection.State)
			}
			ambiguousRegion = ambiguousRegion || request.PeerRegion == nil
		}
	}
	if requireUniqueRegion && ambiguousRegion {
		// Keep omitted-region creation unambiguous for callers such as TGW that
		// still look up connections without a stored effective region.
		return fmt.Errorf("%w for %q: a connection with the same peer account and VPC already exists; specify the peer region explicitly, import the existing connection, or use destroy-before-create when replacing it", ErrAmbiguousPeeringRegion, id)
	}

	_, err = client.VpcPeeringConnectionCreate(ctx, id.Project, id.ProjectVPCID, request)
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

// NewProjectPeeringID builds the API path components from the resource fields.
// The cloud-specific resource adds its region when applicable.
func NewProjectPeeringID(d adapter.ResourceData, accountField, networkField string) (*ProjectPeeringID, error) {
	// vpc_id also supports configurations and SDKv2 state without separate fields.
	var project, vpcID string
	if value, ok := d.GetOk("vpc_id"); ok {
		var err error
		project, vpcID, err = schemautil.SplitResourceID2(value.(string))
		if err != nil {
			return nil, err
		}
	} else {
		project, vpcID = d.Get("project").(string), d.Get("project_vpc_id").(string)
		if project == "" || vpcID == "" {
			return nil, fmt.Errorf("either `vpc_id` or `project` and `project_vpc_id` must be set")
		}
	}
	return &ProjectPeeringID{
		Project: project, ProjectVPCID: vpcID,
		PeerCloudAccount: d.Get(accountField).(string), PeerVPC: d.Get(networkField).(string),
	}, nil
}

// PlanProjectVPC plans project, project_vpc_id and vpc_id from whichever form
// the configuration uses. It replaces the resource only when the Aiven VPC changes:
// schema-level RequiresReplace would also fire when switching forms, because the
// form left out of the configuration is planned as unknown.
func PlanProjectVPC(d adapter.ResourceData) error {
	// Config validators guarantee that one of the forms is set.
	project, projectOK := configString(d, "project")
	vpcID, vpcIDOK := configString(d, "project_vpc_id")
	known := projectOK && vpcIDOK
	if !projectOK && !vpcIDOK {
		if compound, ok := configString(d, "vpc_id"); ok {
			var err error
			project, vpcID, err = schemautil.SplitResourceID2(compound)
			if err != nil {
				return err
			}
			known = true
		}
	}
	// The configured VPC is unknown until apply.
	if !known {
		if !d.IsNewResource() {
			d.RequiresReplace("project", "project_vpc_id", "vpc_id")
		}
		return nil
	}
	for key, value := range map[string]string{
		"project": project, "project_vpc_id": vpcID, "vpc_id": schemautil.BuildResourceID(project, vpcID),
	} {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	if d.IsNewResource() {
		return nil
	}

	stateProject, stateVPCID := d.GetState("project").(string), d.GetState("project_vpc_id").(string)
	if stateProject == "" || stateVPCID == "" {
		// State from before the separate VPC fields, planned without a refresh.
		var err error
		stateProject, stateVPCID, err = schemautil.SplitResourceID2(d.GetState("vpc_id").(string))
		if err != nil {
			return err
		}
		// This migration adds an update after Framework's computed-unknown pass.
		// Computed values may have changed remotely since the last refresh.
		for _, key := range []string{"create_time", "state", "state_info", "aws_vpc_peering_connection_id", "peering_connection_id", "self_link", "user_peer_network_cidrs"} {
			if _, ok := d.Schema().Properties[key]; !ok {
				continue
			}
			if _, configured := d.GetConfigOk(key); configured {
				continue
			}
			if err := d.SetUnknown(key); err != nil {
				return err
			}
		}
	}
	if project != stateProject || vpcID != stateVPCID {
		d.RequiresReplace("project", "project_vpc_id", "vpc_id")
	}
	return nil
}

// configString returns a known, non-null configuration value.
func configString(d adapter.ResourceData, key string) (string, bool) {
	value, ok := d.GetConfigOk(key)
	if !ok {
		return "", false
	}
	return value.(string), true
}

// ProjectPeeringFields fills every form of the Aiven VPC and restores the legacy state_info format.
func ProjectPeeringFields(ctx context.Context, id *ProjectPeeringID, cloud string) adapter.MapModifier {
	return adapter.ComposeMapModifiers(
		PendingPeerWarning(ctx, cloud),
		func(_ adapter.ResourceData, dto map[string]any) error {
			// Legacy AWS/GCP connections can have a null create_time, decoded as time.Time{}.
			// Flatten converts empty strings to Terraform null.
			if dto["create_time"] == (time.Time{}).Format(time.RFC3339Nano) {
				dto["create_time"] = ""
			}
			dto["project"] = id.Project
			dto["project_vpc_id"] = id.ProjectVPCID
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

// formatStateInfo includes all cloud details in a stable order for diagnostics.
func formatStateInfo(info map[string]string) string {
	parts := make([]string, 0, len(info))
	for key, value := range info {
		parts = append(parts, fmt.Sprintf("%s=%q", key, value))
	}
	slices.Sort(parts)
	return strings.Join(parts, ", ")
}

// PeeringCreateStateCheck waits for creation to finish and includes cloud error
// details in diagnostics. Updates only refresh state without checking readiness.
func PeeringCreateStateCheck(d adapter.ResourceData) error {
	if _, exists := d.GetStateOk("id"); exists {
		return nil
	}

	state := d.Get("state").(string)
	detail := formatStateInfo(StateInfoMap(d.Get("state_info").(map[string]any)))
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
