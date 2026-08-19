package vpc

import (
	"testing"

	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestPendingPeerWarning(t *testing.T) {
	t.Run("warns for an existing resource in PENDING_PEER", func(t *testing.T) {
		d := newPendingPeerWarningResourceData(t, adapter.WithTestState(map[string]any{
			"id": "organization/vpc/peer",
		}))

		var got diag.Diagnostics
		ctx, drainWarnings := adapter.WithWarnings(t.Context(), &got)
		err := PendingPeerWarning(ctx, "AWS")(d, map[string]any{
			"state": organizationvpc.VpcPeeringConnectionStateTypePendingPeer,
			"state_info": map[string]any{
				"message": "accept the peering request",
				"type":    "action_required",
			},
		})
		drainWarnings()

		require.NoError(t, err)
		require.Equal(t, 1, got.WarningsCount())
		require.True(t, got.Contains(diag.NewWarningDiagnostic(
			"AWS VPC peering setup is incomplete",
			"Aiven created the peering connection, but it will not become active until the peer setup is completed in AWS. "+
				"State information: accept the peering request\n \"type\":\"action_required\"",
		)))
	})

	t.Run("skips a new resource before refresh", func(t *testing.T) {
		d := newPendingPeerWarningResourceData(t, adapter.WithTestState(map[string]any{"id": ""}))

		var got diag.Diagnostics
		ctx, drainWarnings := adapter.WithWarnings(t.Context(), &got)
		err := PendingPeerWarning(ctx, "AWS")(d, map[string]any{
			"state": organizationvpc.VpcPeeringConnectionStateTypePendingPeer,
		})
		drainWarnings()

		require.NoError(t, err)
		require.Zero(t, got.WarningsCount())
	})

	t.Run("skips a data source", func(t *testing.T) {
		d := newPendingPeerWarningResourceData(
			t,
			adapter.WithTestState(map[string]any{"id": "organization/vpc/peer"}),
			adapter.WithIsDataSource(),
		)

		var got diag.Diagnostics
		ctx, drainWarnings := adapter.WithWarnings(t.Context(), &got)
		err := PendingPeerWarning(ctx, "AWS")(d, map[string]any{
			"state": organizationvpc.VpcPeeringConnectionStateTypePendingPeer,
		})
		drainWarnings()

		require.NoError(t, err)
		require.Zero(t, got.WarningsCount())
	})

	t.Run("skips ACTIVE", func(t *testing.T) {
		d := newPendingPeerWarningResourceData(t, adapter.WithTestState(map[string]any{
			"id": "organization/vpc/peer",
		}))

		var got diag.Diagnostics
		ctx, drainWarnings := adapter.WithWarnings(t.Context(), &got)
		err := PendingPeerWarning(ctx, "AWS")(d, map[string]any{
			"state": organizationvpc.VpcPeeringConnectionStateTypeActive,
		})
		drainWarnings()

		require.NoError(t, err)
		require.Zero(t, got.WarningsCount())
	})
}

func newPendingPeerWarningResourceData(t *testing.T, opts ...adapter.ResourceDataOpt) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		&adapter.Schema{
			Type: adapter.SchemaTypeObject,
			Properties: map[string]*adapter.Schema{
				"id": {Type: adapter.SchemaTypeString, Computed: true},
			},
		},
		[]string{"id"},
		opts...,
	)
	require.NoError(t, err)
	return d
}
