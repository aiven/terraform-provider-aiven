// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package transitgatewayvpcattachment

import (
	"errors"
	"maps"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testProject          = "example-project"
	testProjectVpcID     = "example-vpc"
	testPeerCloudAccount = "123456789012"
	testPeerVpc          = "tgw-0123456789abcdef0"
	testLegacyPeerVpc    = "vpc-0123456789abcdef0"
	testPeerRegion       = "eu-west-3"
	testFourPartID       = "example-project/example-vpc/123456789012/tgw-0123456789abcdef0"
	testFivePartID       = testFourPartID + "/eu-west-3"
	testLegacyFourPartID = "example-project/example-vpc/123456789012/vpc-0123456789abcdef0"
)

func TestParsePeeringID(t *testing.T) {
	f := func(value string, wantRegion *string, wantErr string) {
		t.Helper()

		id, err := parsePeeringID(value)
		if wantErr != "" {
			require.EqualError(t, err, wantErr)
			return
		}

		require.NoError(t, err)
		require.Equal(t, testProject, id.project)
		require.Equal(t, testProjectVpcID, id.projectVpcID)
		require.Equal(t, testPeerCloudAccount, id.peerCloudAccount)
		require.Equal(t, testPeerVpc, id.peerVpc)
		require.Equal(t, wantRegion, id.peerRegion)
	}

	f(testFourPartID, nil, "")
	f(testFivePartID, new(testPeerRegion), "")
	f(testFourPartID+"/", nil, "peer region in the fifth ID component must not be empty")
	f("example-project/example-vpc/123456789012", nil, "expected unix path-like string with 4-5 chunks, got 3")
	f(testFivePartID+"/extra", nil, "expected unix path-like string with 4-5 chunks, got 6")
}

func TestCreateView(t *testing.T) {
	f := func(
		peerRegion, wantAPIPeerRegion *string,
		cidrs []string,
		configureCIDRs bool,
		wantID string,
	) {
		t.Helper()

		ctx := t.Context()
		client := avngen.NewMockClient(t)
		plan := map[string]any{
			"vpc_id":             testProject + "/" + testProjectVpcID,
			"peer_cloud_account": testPeerCloudAccount,
			"peer_vpc":           testPeerVpc,
		}
		if peerRegion != nil {
			plan["peer_region"] = *peerRegion
		}
		if configureCIDRs {
			plan["user_peer_network_cidrs"] = lo.ToAnySlice(cidrs)
		}
		d := newCreateResourceData(t, plan)

		client.EXPECT().
			VpcPeeringConnectionCreate(ctx, testProject, testProjectVpcID, mock.MatchedBy(
				func(req *vpc.VpcPeeringConnectionCreateIn) bool {
					if req.PeerCloudAccount != testPeerCloudAccount || req.PeerVpc != testPeerVpc {
						return false
					}
					if !equalStringPointers(req.PeerRegion, wantAPIPeerRegion) {
						return false
					}
					if len(cidrs) == 0 {
						return req.UserPeerNetworkCidrs == nil
					}
					return req.UserPeerNetworkCidrs != nil && lo.ElementsMatch(*req.UserPeerNetworkCidrs, cidrs)
				},
			)).
			Return(&vpc.VpcPeeringConnectionCreateOut{}, nil).
			Once()

		require.NoError(t, createView(ctx, client, d))
		require.Equal(t, wantID, d.ID())
	}

	f(nil, nil, nil, false, testFourPartID)
	f(new(""), nil, []string{}, true, testFourPartID)
	f(new(testPeerRegion), new(testPeerRegion), []string{"10.10.0.0/24", "10.20.0.0/24"}, true, testFivePartID)
}

func TestSetConnectionStatePeerRegion(t *testing.T) {
	f := func(d adapter.ResourceData, idValue string, wantValue string, wantOK bool) {
		t.Helper()

		id, err := parsePeeringID(idValue)
		require.NoError(t, err)
		connection := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, nil, nil)
		require.NoError(t, setConnectionState(
			d,
			id,
			&connection,
		))

		value, ok := d.GetOk("peer_region")
		require.Equal(t, wantOK, ok)
		if wantOK {
			require.Equal(t, wantValue, value)
		}
	}

	f(newCreateResourceData(t, map[string]any{}), testFourPartID, "", false)
	f(newCreateResourceData(t, map[string]any{"peer_region": ""}), testFourPartID, "", true)
	f(newReadResourceData(t, map[string]any{"id": testFourPartID, "peer_region": ""}), testFourPartID, "", true)
	f(newReadResourceData(t, map[string]any{"id": testFourPartID, "peer_region": testPeerRegion}), testFourPartID, "", false)
	f(newReadResourceData(t, map[string]any{"id": testFivePartID}), testFivePartID, testPeerRegion, true)
}

func TestReadViewReconcilesConfiguredTransitGatewayCIDRs(t *testing.T) {
	desired := []string{"10.20.0.0/24"}
	d := newUpdateResourceData(t, map[string]any{
		"user_peer_network_cidrs": lo.ToAnySlice(desired),
	}, map[string]any{
		"id":                      testFourPartID,
		"user_peer_network_cidrs": []any{"10.10.0.0/24"},
	})
	client := avngen.NewMockClient(t)
	connection := newConnection(
		testPeerRegion,
		vpc.VpcPeeringConnectionStateTypePendingPeer,
		[]string{"10.10.0.0/24"},
		nil,
	)
	expectVpcGet(t, client, connection)
	client.EXPECT().
		VpcPeeringConnectionUpdate(t.Context(), testProject, testProjectVpcID, cidrUpdateMatcher(
			new(testPeerRegion),
			[]string{"10.20.0.0/24"},
			[]string{"10.10.0.0/24"},
		)).
		Return(&vpc.VpcPeeringConnectionUpdateOut{}, nil).
		Once()

	err := readView(t.Context(), client, d)
	require.ErrorIs(t, err, adapter.ErrRefreshStateDesired)
	require.ElementsMatch(t, []any{"10.20.0.0/24"}, d.Get("user_peer_network_cidrs"))
	require.ElementsMatch(t, []any{"10.10.0.0/24"}, d.GetState("user_peer_network_cidrs"))
}

func TestReadViewCIDRReconciliationGuards(t *testing.T) {
	f := func(
		connectionType vpc.VpcPeeringConnectionType,
		state vpc.VpcPeeringConnectionStateType,
		current, configured []string,
		postUpdate bool,
	) {
		t.Helper()

		id := testFourPartID
		peerVpc := testPeerVpc
		if connectionType == vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection {
			id = testLegacyFourPartID
			peerVpc = testLegacyPeerVpc
		}
		plan := map[string]any{"user_peer_network_cidrs": lo.ToAnySlice(configured)}
		var d adapter.ResourceData
		if postUpdate {
			d = newUpdateResourceData(t, plan, map[string]any{
				"id":                      id,
				"user_peer_network_cidrs": lo.ToAnySlice(current),
			})
		} else {
			d = newCreateResourceData(t, plan)
			require.NoError(t, d.SetID(id))
		}
		client := avngen.NewMockClient(t)
		connection := newConnection(testPeerRegion, state, current, nil)
		connection.VpcPeeringConnectionType = connectionType
		connection.PeerVpc = peerVpc
		expectVpcGet(t, client, connection)

		require.NoError(t, readView(t.Context(), client, d))
		require.ElementsMatch(t, lo.ToAnySlice(current), d.Get("user_peer_network_cidrs"))
	}

	f(
		vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
		vpc.VpcPeeringConnectionStateTypeActive,
		[]string{"10.10.0.0/24", "10.20.0.0/24"},
		[]string{"10.20.0.0/24", "10.10.0.0/24"},
		true,
	)
	f(
		vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection,
		vpc.VpcPeeringConnectionStateTypeActive,
		[]string{},
		[]string{"10.20.0.0/24"},
		false,
	)
	f(
		vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
		vpc.VpcPeeringConnectionStateTypeApprovedPeerRequested,
		[]string{},
		[]string{"10.20.0.0/24"},
		false,
	)
}

func TestReadViewDoesNotReconcileCIDRsWithoutConfiguration(t *testing.T) {
	d := newReadResourceData(t, map[string]any{
		"id":                      testFourPartID,
		"user_peer_network_cidrs": []any{"10.10.0.0/24"},
	})
	client := avngen.NewMockClient(t)
	expectVpcGet(t, client, newConnection(
		testPeerRegion,
		vpc.VpcPeeringConnectionStateTypeActive,
		[]string{},
		nil,
	))

	require.NoError(t, readView(t.Context(), client, d))
	value, ok := d.GetOk("user_peer_network_cidrs")
	require.True(t, ok)
	require.Empty(t, value)
}

func TestReadViewAddsPendingPeerWarning(t *testing.T) {
	d := newReadResourceData(t, map[string]any{"id": testFivePartID})
	client := avngen.NewMockClient(t)
	expectVpcGet(t, client, newConnection(
		testPeerRegion,
		vpc.VpcPeeringConnectionStateTypePendingPeer,
		nil,
		map[string]any{
			"aws_transit_gateway_attachment_id": "tgw-attach-123",
			"message":                           "accept the attachment",
			"type":                              "action-required",
		},
	))

	var diagnostics diag.Diagnostics
	ctx, drainWarnings := adapter.WithWarnings(t.Context(), &diagnostics)
	require.NoError(t, readView(ctx, client, d))
	drainWarnings()

	require.Equal(t, 1, diagnostics.WarningsCount())
	require.True(t, diagnostics.Contains(diag.NewWarningDiagnostic(
		"VPC peering connection is pending peer setup",
		"Aiven created its side of the connection, but the connection isn't active until the setup is completed in the peer cloud account. "+
			`State info: aws_transit_gateway_attachment_id="tgw-attach-123", message="accept the attachment", type="action-required"`,
	)))
}

func TestStateInfoMap(t *testing.T) {
	f := func(info map[string]any, want map[string]string) {
		t.Helper()

		require.Equal(t, want, stateInfoMap(info))
	}

	f(nil, nil)
	f(map[string]any{
		"aws_transit_gateway_attachment_id": "tgw-attach-123",
		"empty":                             "",
		"future_nested":                     map[string]any{"details": []any{"alpha", 42}},
		"message":                           "attachment available",
		"type":                              "attachment-ready",
	}, map[string]string{
		"aws_transit_gateway_attachment_id": "tgw-attach-123",
		"empty":                             "",
		"future_nested":                     "map[details:[alpha 42]]",
		"message":                           "attachment available",
		"type":                              "attachment-ready",
	})
}

func TestSetConnectionStateStateInfo(t *testing.T) {
	f := func(
		connectionType vpc.VpcPeeringConnectionType,
		info map[string]any,
		wantStateInfo map[string]any,
		wantProviderID string,
	) {
		t.Helper()

		idValue := testFourPartID
		peerVpc := testPeerVpc
		if connectionType == vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection {
			idValue = testLegacyFourPartID
			peerVpc = testLegacyPeerVpc
		}
		connectionState := vpc.VpcPeeringConnectionStateTypeActive
		if connectionType == vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection {
			connectionState = vpc.VpcPeeringConnectionStateTypePendingPeer
		}
		d := newReadResourceData(t, map[string]any{
			"id":                    idValue,
			"peering_connection_id": "pcx-stale",
			"state_info": map[string]any{
				"aws_vpc_peering_connection_id": "pcx-stale",
				"obsolete":                      "stale-value",
			},
		})
		connection := newConnection(
			testPeerRegion,
			connectionState,
			[]string{},
			info,
		)
		connection.VpcPeeringConnectionType = connectionType
		connection.PeerVpc = peerVpc
		id, err := parsePeeringID(idValue)
		require.NoError(t, err)

		require.NoError(t, setConnectionState(d, id, &connection))
		if wantStateInfo == nil {
			require.Empty(t, d.Get("state_info"))
		} else {
			require.Equal(t, wantStateInfo, d.Get("state_info"))
		}
		providerID, hasProviderID := d.GetOk("peering_connection_id")
		if wantProviderID == "" {
			require.False(t, hasProviderID)
		} else {
			require.True(t, hasProviderID)
			require.Equal(t, wantProviderID, providerID)
		}
		require.Equal(t, testProject+"/"+testProjectVpcID, d.Get("vpc_id"))
		require.Equal(t, testPeerCloudAccount, d.Get("peer_cloud_account"))
		require.Equal(t, peerVpc, d.Get("peer_vpc"))
		require.Equal(t, string(connectionState), d.Get("state"))
		cidrs, ok := d.GetOk("user_peer_network_cidrs")
		require.True(t, ok)
		require.Empty(t, cidrs)
	}

	f(vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection, map[string]any{
		"aws_vpc_peering_connection_id": "pcx-fresh",
		"message":                       "accept the peering request",
		"type":                          "aws-vpc-peering-connection-pending-peer",
	}, map[string]any{
		"aws_vpc_peering_connection_id": "pcx-fresh",
		"message":                       "accept the peering request",
		"type":                          "aws-vpc-peering-connection-pending-peer",
	}, "pcx-fresh")
	f(vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection, map[string]any{
		"aws_vpc_peering_connection_id": "",
		"message":                       "waiting for peer",
		"type":                          "aws-vpc-peering-connection-pending-peer",
	}, map[string]any{
		"aws_vpc_peering_connection_id": "",
		"message":                       "waiting for peer",
		"type":                          "aws-vpc-peering-connection-pending-peer",
	}, "")
	f(vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment, map[string]any{
		"aws_transit_gateway_attachment_id": "tgw-attach-fresh",
		"message":                           "attachment available",
		"type":                              "aws-transit-gateway-vpc-attachment-available",
	}, map[string]any{
		"aws_transit_gateway_attachment_id": "tgw-attach-fresh",
		"message":                           "attachment available",
		"type":                              "aws-transit-gateway-vpc-attachment-available",
	}, "")
	f(vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment, nil, nil, "")
}

func TestFindPeeringConnection(t *testing.T) {
	f := func(
		idValue string,
		connections []vpc.PeeringConnectionOut,
		wantCIDR string,
		wantErr error,
	) {
		t.Helper()

		id, err := parsePeeringID(idValue)
		require.NoError(t, err)
		client := avngen.NewMockClient(t)
		expectVpcGet(t, client, connections...)
		connection, err := findPeeringConnection(t.Context(), client, id)
		if wantErr != nil {
			require.ErrorIs(t, err, wantErr)
			return
		}

		require.NoError(t, err)
		require.Equal(t, []string{wantCIDR}, connection.UserPeerNetworkCidrs)
	}

	wrong := newConnection("us-east-1", vpc.VpcPeeringConnectionStateTypeActive, []string{"10.1.0.0/24"}, nil)
	correct := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, []string{"10.2.0.0/24"}, nil)
	wrongAccount := correct
	wrongAccount.PeerCloudAccount = "000000000000"
	wrongVpc := correct
	wrongVpc.PeerVpc = "tgw-fffffffffffffffff"
	missingRegion := correct
	missingRegion.PeerRegion = nil
	emptyRegion := correct
	emptyRegion.PeerRegion = new("")
	f(testFourPartID, []vpc.PeeringConnectionOut{wrong}, "10.1.0.0/24", nil)
	f(testFourPartID, []vpc.PeeringConnectionOut{wrong, correct}, "", adapter.ErrMultiple)
	f(
		testFivePartID,
		[]vpc.PeeringConnectionOut{wrong, wrongAccount, wrongVpc, missingRegion, emptyRegion, correct},
		"10.2.0.0/24",
		nil,
	)
}

func TestDatasourceReadView(t *testing.T) {
	f := func(
		requestedRegion *string,
		connections []vpc.PeeringConnectionOut,
		wantID string,
		wantCIDR string,
		wantRegion *string,
	) {
		t.Helper()

		d := newDatasourceData(t, requestedRegion)
		client := avngen.NewMockClient(t)
		expectVpcGet(t, client, connections...)
		require.NoError(t, DataSourceOptions.Read(t.Context(), client, d))
		require.Equal(t, wantID, d.ID())
		require.ElementsMatch(t, []any{wantCIDR}, d.Get("user_peer_network_cidrs"))
		switch {
		case wantRegion == nil:
			_, ok := d.GetOk("peer_region")
			require.False(t, ok)
		case *wantRegion == "":
			peerRegion, ok := d.GetOk("peer_region")
			require.True(t, ok)
			require.Empty(t, peerRegion)
		default:
			peerRegion, ok := d.GetOk("peer_region")
			require.True(t, ok)
			require.Equal(t, *wantRegion, peerRegion)
		}
	}

	west := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, []string{"10.2.0.0/24"}, nil)
	east := newConnection("us-east-1", vpc.VpcPeeringConnectionStateTypeActive, []string{"10.1.0.0/24"}, nil)
	regionless := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, []string{"10.3.0.0/24"}, nil)
	regionless.PeerRegion = nil
	f(new(testPeerRegion), []vpc.PeeringConnectionOut{east, west}, testFivePartID, "10.2.0.0/24", new(testPeerRegion))
	f(nil, []vpc.PeeringConnectionOut{west}, testFivePartID, "10.2.0.0/24", new(testPeerRegion))
	f(new(""), []vpc.PeeringConnectionOut{west}, testFivePartID, "10.2.0.0/24", new(""))
	f(nil, []vpc.PeeringConnectionOut{regionless}, testFourPartID, "10.3.0.0/24", nil)
}

func TestModifyPlanCIDRs(t *testing.T) {
	f := func(state string, prior, planned, want []string) {
		t.Helper()

		d, err := adapter.NewResourceData(
			resourceSchemaInternal(),
			idFields(),
			adapter.WithTestPlan(map[string]any{
				"user_peer_network_cidrs": lo.ToAnySlice(planned),
			}),
			adapter.WithTestState(map[string]any{
				"id":                      testFourPartID,
				"state":                   state,
				"user_peer_network_cidrs": lo.ToAnySlice(prior),
			}),
			adapter.WithTestConfig(map[string]any{
				"user_peer_network_cidrs": lo.ToAnySlice(planned),
			}),
		)
		require.NoError(t, err)
		require.NoError(t, modifyPlan(t.Context(), nil, d))
		got := lo.Must(lo.FromAnySlice[string](d.Get("user_peer_network_cidrs").([]any)))
		require.ElementsMatch(t, want, got)
	}

	configured := []string{"10.20.0.0/24"}
	empty := []string{}
	f("APPROVED", empty, configured, configured)
	f("ACTIVE", empty, configured, configured)
	f("PENDING_PEER", empty, configured, configured)
	f("APPROVED_PEER_REQUESTED", empty, configured, empty)

	// A create, including Terraform's second planning pass for a replacement,
	// has no prior ID and must retain the configured CIDRs.
	d := newCreateResourceData(t, map[string]any{
		"user_peer_network_cidrs": lo.ToAnySlice(configured),
	})
	require.NoError(t, modifyPlan(t.Context(), nil, d))
	got := lo.Must(lo.FromAnySlice[string](d.Get("user_peer_network_cidrs").([]any)))
	require.ElementsMatch(t, configured, got)

	unknown := tftypes.NewValue(
		tftypes.Set{ElementType: tftypes.String},
		tftypes.UnknownValue,
	)
	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(map[string]any{
			"user_peer_network_cidrs": unknown,
		}),
		adapter.WithTestState(map[string]any{
			"id":                      testFourPartID,
			"state":                   "REJECTED_BY_PEER",
			"user_peer_network_cidrs": lo.ToAnySlice(empty),
		}),
		adapter.WithTestConfig(map[string]any{
			"user_peer_network_cidrs": unknown,
		}),
	)
	require.NoError(t, err)
	require.NoError(t, modifyPlan(t.Context(), nil, d))
	_, known := d.GetOk("user_peer_network_cidrs")
	require.False(t, known, "an unknown planned value must not be replaced with prior state")
}

func TestUpdateViewUsesExactRegionAndCIDRDelta(t *testing.T) {
	ctx := t.Context()
	client := avngen.NewMockClient(t)
	d := newUpdateResourceData(t, map[string]any{
		"user_peer_network_cidrs": []any{"10.20.0.0/24"},
	}, map[string]any{
		"id":                      testFivePartID,
		"user_peer_network_cidrs": []any{"10.10.0.0/24"},
	})
	client.EXPECT().
		VpcGet(ctx, testProject, testProjectVpcID).
		Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{
				{
					PeerCloudAccount:         testPeerCloudAccount,
					PeerRegion:               new("us-east-1"),
					PeerVpc:                  testPeerVpc,
					State:                    vpc.VpcPeeringConnectionStateTypeActive,
					UserPeerNetworkCidrs:     []string{"10.1.0.0/24"},
					VpcPeeringConnectionType: vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
				},
				{
					PeerCloudAccount:         testPeerCloudAccount,
					PeerRegion:               new(testPeerRegion),
					PeerVpc:                  testPeerVpc,
					State:                    vpc.VpcPeeringConnectionStateTypeApproved,
					UserPeerNetworkCidrs:     []string{"10.10.0.0/24"},
					VpcPeeringConnectionType: vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
				},
			},
		}, nil).
		Once()
	client.EXPECT().
		VpcPeeringConnectionUpdate(ctx, testProject, testProjectVpcID, cidrUpdateMatcher(
			new(testPeerRegion),
			[]string{"10.20.0.0/24"},
			[]string{"10.10.0.0/24"},
		)).
		Return(&vpc.VpcPeeringConnectionUpdateOut{}, nil).
		Once()

	require.NoError(t, updateView(ctx, client, d))
}

func TestUpdateViewCIDRGuards(t *testing.T) {
	f := func(
		connectionType vpc.VpcPeeringConnectionType,
		state vpc.VpcPeeringConnectionStateType,
		current, desired []string,
		wantErr string,
	) {
		t.Helper()

		ctx := t.Context()
		client := avngen.NewMockClient(t)
		id := testFivePartID
		peerVpc := testPeerVpc
		if connectionType == vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection {
			id = testLegacyFourPartID + "/" + testPeerRegion
			peerVpc = testLegacyPeerVpc
		}
		d := newUpdateResourceData(t, map[string]any{
			"user_peer_network_cidrs": lo.ToAnySlice(desired),
		}, map[string]any{
			"id":                      id,
			"user_peer_network_cidrs": lo.ToAnySlice(current),
		})
		client.EXPECT().
			VpcGet(ctx, testProject, testProjectVpcID).
			Return(&vpc.VpcGetOut{PeeringConnections: []vpc.PeeringConnectionOut{
				{
					PeerCloudAccount:         testPeerCloudAccount,
					PeerRegion:               new(testPeerRegion),
					PeerVpc:                  peerVpc,
					State:                    state,
					UserPeerNetworkCidrs:     current,
					VpcPeeringConnectionType: connectionType,
				},
			}}, nil).
			Once()

		err := updateView(ctx, client, d)
		if wantErr == "" {
			require.NoError(t, err)
			return
		}
		require.ErrorContains(t, err, wantErr)
	}

	f(
		vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection,
		vpc.VpcPeeringConnectionStateTypeActive,
		[]string{"10.10.0.0/24"},
		[]string{"10.20.0.0/24"},
		"cannot update user_peer_network_cidrs for VPC peering connection type",
	)
	f(
		vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
		vpc.VpcPeeringConnectionStateTypeRejectedByPeer,
		[]string{},
		[]string{"10.20.0.0/24"},
		"cannot update user_peer_network_cidrs while VPC peering connection is in state",
	)
	f(
		vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
		vpc.VpcPeeringConnectionStateTypeRejectedByPeer,
		[]string{},
		[]string{},
		"",
	)
}

func TestRefreshStateCheck(t *testing.T) {
	f := func(state string, stateInfo map[string]any, wantErr, wantFailed bool, wantContains string) {
		t.Helper()

		values := map[string]any{"state": state}
		if stateInfo != nil {
			values["state_info"] = stateInfo
		}
		d := newReadResourceData(t, values)
		err := refreshStateCheck(d)
		if !wantErr {
			require.NoError(t, err)
			return
		}

		require.Error(t, err)
		require.Equal(t, wantFailed, errors.Is(err, adapter.ErrRefreshStateFailed))
		require.ErrorContains(t, err, wantContains)
	}

	f("ACTIVE", nil, false, false, "")
	f("PENDING_PEER", nil, false, false, "")
	f("APPROVED", map[string]any{
		"type":    "action-required",
		"message": "backend detail",
	}, true, false, `state_info: message="backend detail", type="action-required"`)
	f("APPROVED_PEER_REQUESTED", nil, true, false, "transient state")
	f("DELETED", nil, true, true, "was deleted")
	f("DELETING", nil, true, true, "was deleted")
	f("DELETED_BY_PEER", nil, true, true, "peer cloud resource was deleted")
	f("REJECTED_BY_PEER", nil, true, true, "rejected by the peer")
	f("INVALID_SPECIFICATION", nil, true, true, "specification is invalid")
	f("ERROR", nil, true, true, "reached ERROR")
	f("FUTURE_STATE", nil, true, false, `unknown VPC peering connection state "FUTURE_STATE"`)
}

func TestDeleteViewSelectsEndpointFromID(t *testing.T) {
	f := func(id string) {
		t.Helper()

		ctx := t.Context()
		client := avngen.NewMockClient(t)
		d := newReadResourceData(t, map[string]any{"id": id})
		if id == testFourPartID {
			client.EXPECT().
				VpcPeeringConnectionDelete(ctx, testProject, testProjectVpcID, testPeerCloudAccount, testPeerVpc).
				Return(&vpc.VpcPeeringConnectionDeleteOut{}, nil).
				Once()
		} else {
			client.EXPECT().
				VpcPeeringConnectionWithRegionDelete(
					ctx,
					testProject,
					testProjectVpcID,
					testPeerCloudAccount,
					testPeerVpc,
					testPeerRegion,
				).
				Return(&vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil).
				Once()
		}

		require.NoError(t, deleteView(ctx, client, d))
	}

	f(testFourPartID)
	f(testFivePartID)
}

func newCreateResourceData(t *testing.T, plan map[string]any) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(plan),
		adapter.WithTestConfig(maps.Clone(plan)),
	)
	require.NoError(t, err)
	return d
}

func newUpdateResourceData(t *testing.T, plan, state map[string]any) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestPlan(plan),
		adapter.WithTestState(state),
		adapter.WithTestConfig(maps.Clone(plan)),
	)
	require.NoError(t, err)
	return d
}

func newReadResourceData(t *testing.T, state map[string]any) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)
	return d
}

func newDatasourceData(t *testing.T, peerRegion *string) adapter.ResourceData {
	t.Helper()

	config := map[string]any{
		"vpc_id":             testProject + "/" + testProjectVpcID,
		"peer_cloud_account": testPeerCloudAccount,
		"peer_vpc":           testPeerVpc,
	}
	if peerRegion != nil {
		config["peer_region"] = *peerRegion
	}
	d, err := adapter.NewResourceData(
		datasourceSchemaInternal(),
		idFields(),
		adapter.WithIsDataSource(),
		adapter.WithTestConfig(config),
	)
	require.NoError(t, err)
	return d
}

func newConnection(
	region string,
	state vpc.VpcPeeringConnectionStateType,
	cidrs []string,
	stateInfo map[string]any,
) vpc.PeeringConnectionOut {
	return vpc.PeeringConnectionOut{
		PeerCloudAccount:         testPeerCloudAccount,
		PeerRegion:               new(region),
		PeerVpc:                  testPeerVpc,
		State:                    state,
		StateInfo:                stateInfo,
		UserPeerNetworkCidrs:     cidrs,
		VpcPeeringConnectionType: vpc.VpcPeeringConnectionTypeAWSTgwVpcAttachment,
	}
}

func expectVpcGet(
	t *testing.T,
	client *avngen.MockClient,
	connections ...vpc.PeeringConnectionOut,
) {
	t.Helper()

	client.EXPECT().
		VpcGet(mock.Anything, testProject, testProjectVpcID).
		Return(&vpc.VpcGetOut{
			PeeringConnections: connections,
		}, nil).
		Once()
}

func cidrUpdateMatcher(peerRegion *string, add, remove []string) any {
	return mock.MatchedBy(func(req *vpc.VpcPeeringConnectionUpdateIn) bool {
		if req.Add == nil || req.Delete == nil || !lo.ElementsMatch(*req.Delete, remove) {
			return false
		}
		gotAdd := make([]string, 0, len(*req.Add))
		for _, item := range *req.Add {
			if item.PeerCloudAccount != testPeerCloudAccount ||
				item.PeerVpc != testPeerVpc ||
				!equalStringPointers(item.PeerRegion, peerRegion) {
				return false
			}
			if item.PeerResourceGroup != nil {
				return false
			}
			gotAdd = append(gotAdd, item.Cidr)
		}
		return lo.ElementsMatch(gotAdd, add)
	})
}

func equalStringPointers(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
