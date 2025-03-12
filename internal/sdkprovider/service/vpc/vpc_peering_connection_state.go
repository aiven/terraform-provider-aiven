package vpc

import (
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// this code provides adapter implementations for VPC peering connection state handling
// between different HTTP clients (aiven-go-client and go-client-codegen). It defines a common
// interface peeringConnectionState and wrapper types that adapt different client responses to
// this interface, enabling state checks and diagnostics without major code refactoring.
type peeringConnectionState interface {
	GetState() string
	GetStateInfo() *map[string]any
}

func getDiagnosticsFromState(pc peeringConnectionState) diag.Diagnostics {
	state := organizationvpc.VpcPeeringConnectionStateType(pc.GetState())

	switch state {
	case organizationvpc.VpcPeeringConnectionStateTypeActive:
		return nil
	case organizationvpc.VpcPeeringConnectionStateTypePendingPeer:
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary: fmt.Sprintf("Aiven platform has created a connection to the specified "+
				"peer successfully in the cloud, but the connection is not active until the user "+
				"completes the setup in their cloud account. The steps needed in the user cloud "+
				"account depend on the used cloud provider. Find more in the state info: %s",
				stateInfoToString(pc.GetStateInfo()))}}
	case organizationvpc.VpcPeeringConnectionStateTypeDeleted, organizationvpc.VpcPeeringConnectionStateTypeDeleting:
		return diag.Errorf("A user has deleted the peering connection through the Aiven " +
			"Terraform provider, or Aiven Web Console or directly via Aiven API. There are no " +
			"transitions from this state")
	case organizationvpc.VpcPeeringConnectionStateTypeDeletedByPeer:
		return diag.Errorf("A user deleted the peering cloud resource in their account. " +
			"There are no transitions from this state")
	case organizationvpc.VpcPeeringConnectionStateTypeRejectedByPeer:
		return diag.Errorf("VPC peering connection request was rejected, state info: %s",
			stateInfoToString(pc.GetStateInfo()))
	case organizationvpc.VpcPeeringConnectionStateTypeInvalidSpecification:
		return diag.Errorf("VPC peering connection cannot be created, more in the state info: %s",
			stateInfoToString(pc.GetStateInfo()))
	default:
		return diag.Errorf("Unknown VPC peering connection state: %s", pc.GetState())
	}
}

// stateInfoToString converts VPC peering connection state_info to a string
func stateInfoToString(s *map[string]interface{}) string {
	if s == nil || len(*s) == 0 {
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

type aivenVPCPeeringWrapper struct {
	*aiven.VPCPeeringConnection
}

// Create wrapper functions instead of methods
func newAivenVPCPeeringState(pc *aiven.VPCPeeringConnection) peeringConnectionState {
	return &aivenVPCPeeringWrapper{pc}
}

func (w *aivenVPCPeeringWrapper) GetState() string {
	return w.State
}

func (w *aivenVPCPeeringWrapper) GetStateInfo() *map[string]any {
	return w.StateInfo
}

type organizationVPCPeeringWrapper struct {
	*organizationvpc.OrganizationVpcGetPeeringConnectionOut
}

func newOrganizationVPCPeeringState(pc *organizationvpc.OrganizationVpcGetPeeringConnectionOut) *organizationVPCPeeringWrapper {
	return &organizationVPCPeeringWrapper{pc}
}

func (w *organizationVPCPeeringWrapper) GetState() string {
	return string(w.State)
}

func (w *organizationVPCPeeringWrapper) GetStateInfo() *map[string]any {
	stateInfo := make(map[string]any)

	stateInfo["message"] = w.StateInfo.Message
	stateInfo["type"] = w.StateInfo.Type

	return &stateInfo
}
