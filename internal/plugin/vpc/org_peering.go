package vpc

import (
	"context"
	"fmt"

	"github.com/aiven/go-client-codegen/handler/organizationvpc"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// PendingPeerWarning warns that the cloud-side VPC peering setup still needs to be completed.
// A newly created resource is skipped because its initial Flatten happens before refreshState;
// the subsequent Read reports the final state and emits the warning when needed.
func PendingPeerWarning(ctx context.Context, cloud string) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		if !d.IsResource() || d.IsNewResource() || !adapter.Equal(dto["state"], organizationvpc.VpcPeeringConnectionStateTypePendingPeer) {
			return nil
		}

		detail := fmt.Sprintf("Aiven created the peering connection, but it will not become active until the peer setup is completed in %s.", cloud)
		if stateInfo, ok := dto["state_info"].(map[string]any); ok {
			message, _ := stateInfo["message"].(string)
			stateType, _ := stateInfo["type"].(string)
			if stateType != "" {
				if message != "" {
					message += "\n "
				}
				message += fmt.Sprintf("%q:%q", "type", stateType)
			}
			if message != "" {
				detail += " State information: " + message
			}
		}

		adapter.AddWarning(ctx, fmt.Sprintf("%s VPC peering setup is incomplete", cloud), detail)
		return nil
	}
}
