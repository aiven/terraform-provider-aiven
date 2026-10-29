// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package transitgatewayvpcattachment

import (
	"context"
	"maps"
	"testing"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
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
		require.Equal(t, testProject, id.Project)
		require.Equal(t, testProjectVpcID, id.ProjectVPCID)
		require.Equal(t, testPeerCloudAccount, id.PeerCloudAccount)
		require.Equal(t, testPeerVpc, id.PeerVPC)
		require.Equal(t, wantRegion, id.PeerRegion)
	}
	f(testFourPartID, nil, "")
	f(testFivePartID, new(testPeerRegion), "")
	f(testFourPartID+"/", nil, "peer region in the fifth ID component must not be empty")
	f("example-project/example-vpc/123456789012", nil, "expected unix path-like string with 4-5 chunks, got 3")
	f(testFivePartID+"/extra", nil, "expected unix path-like string with 4-5 chunks, got 6")
}

func TestCreateView(t *testing.T) {
	f := func(
		peerRegion, wantIDRegion *string,
		cidrs []string,
		configureCIDRs bool,
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

		client.EXPECT().VpcGet(ctx, testProject, testProjectVpcID).
			Return(&vpc.VpcGetOut{CloudName: "aws-" + testPeerRegion}, nil).Once()
		client.EXPECT().
			VpcPeeringConnectionCreate(ctx, testProject, testProjectVpcID, mock.MatchedBy(
				func(req *vpc.VpcPeeringConnectionCreateIn) bool {
					if req.PeerCloudAccount != testPeerCloudAccount || req.PeerVpc != testPeerVpc {
						return false
					}
					// The API resolves omitted regions; the provider sends explicit ones unchanged.
					if !equalStringPointers(req.PeerRegion, wantIDRegion) {
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
		wantID := testFourPartID
		if wantIDRegion != nil {
			wantID += "/" + *wantIDRegion
		}
		require.Equal(t, wantID, d.ID())
	}

	f(nil, nil, nil, false)
	f(new(""), nil, []string{}, true)
	f(new(testPeerRegion), new(testPeerRegion), []string{"10.10.0.0/24", "10.20.0.0/24"}, true)
}

func TestCreateViewRejectsAmbiguousRegion(t *testing.T) {
	client := avngen.NewMockClient(t)
	d := newCreateResourceData(t, map[string]any{
		"vpc_id":             testProject + "/" + testProjectVpcID,
		"peer_cloud_account": testPeerCloudAccount, "peer_vpc": testLegacyPeerVpc,
	})
	other := newConnection("us-east-1", vpc.VpcPeeringConnectionStateTypeInvalidSpecification, nil, nil)
	other.PeerVpc = testLegacyPeerVpc
	other.VpcPeeringConnectionType = vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection
	client.EXPECT().VpcGet(t.Context(), testProject, testProjectVpcID).Return(&vpc.VpcGetOut{
		CloudName: "aws-" + testPeerRegion, PeeringConnections: []vpc.PeeringConnectionOut{other},
	}, nil).Once()
	require.ErrorIs(t, createView(t.Context(), client, d), pluginvpc.ErrAmbiguousPeeringRegion)
	client.AssertNumberOfCalls(t, "VpcPeeringConnectionCreate", 0)
}

func TestSetConnectionStatePeerRegion(t *testing.T) {
	f := func(d adapter.ResourceData, id *pluginvpc.ProjectPeeringID, wantValue string, wantOK bool) {
		t.Helper()

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
		require.Equal(t, id.String(), d.ID())
	}
	read := func(state map[string]any) (adapter.ResourceData, *pluginvpc.ProjectPeeringID) {
		d := newReadResourceData(t, state)
		id, err := parsePeeringID(d.ID())
		require.NoError(t, err)
		return d, id
	}

	f(newCreateResourceData(t, map[string]any{}), parseID(t, testFourPartID), "", false)
	f(newCreateResourceData(t, map[string]any{"peer_region": ""}), parseID(t, testFourPartID), "", true)
	d, id := read(map[string]any{"id": testFourPartID, "peer_region": ""})
	f(d, id, "", true)
	d, id = read(map[string]any{"id": testFourPartID, "peer_region": testPeerRegion})
	f(d, id, "", false)
	d, id = read(map[string]any{"id": testFourPartID})
	f(d, id, "", false)
	d, id = read(map[string]any{"id": testFivePartID})
	f(d, id, testPeerRegion, true)
}

func TestImportPreservesID(t *testing.T) {
	check := func(t *testing.T, idValue, wantRegion string) {
		t.Helper()
		ctx := t.Context()
		schema := resourceSchema(ctx)
		rsp := resource.ImportStateResponse{State: tfsdk.State{
			Schema: schema,
			Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		}}
		adapter.NewResource(ResourceOptions).(resource.ResourceWithImportState).ImportState(ctx,
			resource.ImportStateRequest{ID: idValue}, &rsp)
		require.False(t, rsp.Diagnostics.HasError(), "%v", rsp.Diagnostics)
		d, err := adapter.NewResourceData(resourceSchemaInternal(), idFields(), adapter.WithState(rsp.State))
		require.NoError(t, err)
		require.Equal(t, idValue, d.ID())

		client := avngen.NewMockClient(t)
		expectVpcGet(t, client, newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, nil, nil))
		require.NoError(t, readView(ctx, client, d))
		require.Equal(t, idValue, d.ID())
		require.Equal(t, wantRegion, d.Get("peer_region"))
		require.Equal(t, testProject, d.Get("project"))
		require.Equal(t, testProjectVpcID, d.Get("project_vpc_id"))
	}
	t.Run("four parts", func(t *testing.T) {
		check(t, testFourPartID, "")
	})
	t.Run("five parts", func(t *testing.T) {
		check(t, testFivePartID, testPeerRegion)
	})
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
	require.NoError(t, err, "PUT persists CIDRs before returning; another refresh is unnecessary")
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
		"VPC peering setup is incomplete",
		"Aiven created the peering connection, but it will not become active until the peer setup is completed in the peer cloud account. "+
			"State information: accept the attachment\n \"type\":\"action-required\"",
	)))
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

		require.NoError(t, setConnectionState(d, parseID(t, idValue), &connection))
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

func TestReadViewSelectsConnection(t *testing.T) {
	f := func(
		idValue string,
		connections []vpc.PeeringConnectionOut,
		wantCIDR string,
		wantErr error,
	) {
		t.Helper()

		d := newReadResourceData(t, map[string]any{"id": idValue})
		client := avngen.NewMockClient(t)
		expectVpcGet(t, client, connections...)
		err := readView(t.Context(), client, d)
		if wantErr != nil {
			require.ErrorIs(t, err, wantErr)
			return
		}

		require.NoError(t, err)
		require.Equal(t, []any{wantCIDR}, d.Get("user_peer_network_cidrs"))
		require.Equal(t, idValue, d.ID())

		// A second refresh must still select the same connection and preserve its ID.
		expectVpcGet(t, client, connections...)
		require.NoError(t, readView(t.Context(), client, d))
		require.Equal(t, []any{wantCIDR}, d.Get("user_peer_network_cidrs"))
		require.Equal(t, idValue, d.ID())
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
		wantCIDR string,
		wantRegion *string,
		wantID string,
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
	f(new(testPeerRegion), []vpc.PeeringConnectionOut{east, west}, "10.2.0.0/24", new(testPeerRegion), testFivePartID)
	f(nil, []vpc.PeeringConnectionOut{west}, "10.2.0.0/24", new(testPeerRegion), testFivePartID)
	f(new(""), []vpc.PeeringConnectionOut{west}, "10.2.0.0/24", new(""), testFivePartID)
	f(nil, []vpc.PeeringConnectionOut{regionless}, "10.3.0.0/24", nil, testFourPartID)
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

func TestModifyPlanRejectsCIDRsForVPCPeering(t *testing.T) {
	check := func(t *testing.T, peer any, cidrs any, wantError bool) {
		t.Helper()
		config := map[string]any{"peer_vpc": peer}
		if cidrs != nil {
			config["user_peer_network_cidrs"] = cidrs
		}
		d := newCreateResourceData(t, config)
		err := modifyPlan(t.Context(), nil, d)
		if wantError {
			require.ErrorContains(t, err, "user_peer_network_cidrs")
		} else {
			require.NoError(t, err)
		}
	}
	t.Run("ordinary peering with CIDRs", func(t *testing.T) {
		check(t, testLegacyPeerVpc, []any{"10.20.0.0/24"}, true)
	})
	t.Run("transit gateway with CIDRs", func(t *testing.T) {
		check(t, testPeerVpc, []any{"10.20.0.0/24"}, false)
	})
	t.Run("empty CIDRs", func(t *testing.T) {
		check(t, testLegacyPeerVpc, []any{}, false)
	})
	t.Run("omitted CIDRs", func(t *testing.T) {
		check(t, testLegacyPeerVpc, nil, false)
	})
	t.Run("null CIDRs", func(t *testing.T) {
		check(t, testLegacyPeerVpc, tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil), false)
	})
	t.Run("unknown CIDRs", func(t *testing.T) {
		check(t, testLegacyPeerVpc, tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue), false)
	})
	t.Run("unknown peer", func(t *testing.T) {
		check(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), []any{"10.20.0.0/24"}, false)
	})
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

func TestDeleteViewSelectsEndpointFromRegion(t *testing.T) {
	f := func(t *testing.T, state map[string]any, region *string) {
		t.Helper()

		ctx := t.Context()
		client := avngen.NewMockClient(t)
		d := newReadResourceData(t, state)
		if state["id"] == testFourPartID {
			connection := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, nil, nil)
			connection.PeerRegion = region
			expectVpcGet(t, client, connection)
		}
		if region == nil || *region == "" {
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

	t.Run("explicit region", func(t *testing.T) {
		f(t, map[string]any{"id": testFivePartID}, new(testPeerRegion))
	})
	t.Run("resolve region despite stale state", func(t *testing.T) {
		f(t, map[string]any{"id": testFourPartID, "peer_region": "us-east-1"}, new(testPeerRegion))
	})
	t.Run("null region", func(t *testing.T) {
		f(t, map[string]any{"id": testFourPartID}, nil)
	})
	t.Run("empty region", func(t *testing.T) {
		f(t, map[string]any{"id": testFourPartID}, new(""))
	})
	t.Run("missing connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		expectVpcGet(t, client)
		d := newReadResourceData(t, map[string]any{"id": testFourPartID})
		require.ErrorIs(t, deleteView(t.Context(), client, d), adapter.ErrNotFound)
	})
	t.Run("ambiguous connection", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		expectVpcGet(t, client,
			newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, nil, nil),
			newConnection("us-east-1", vpc.VpcPeeringConnectionStateTypeInvalidSpecification, nil, nil),
		)
		d := newReadResourceData(t, map[string]any{"id": testFourPartID})
		require.ErrorIs(t, deleteView(t.Context(), client, d), adapter.ErrMultiple)
	})
}

func TestMigrationWithoutRefreshRejectedConnection(t *testing.T) {
	sch := resourceSchema(t.Context())
	values := map[string]any{
		"vpc_id": testProject + "/" + testProjectVpcID, "peer_cloud_account": testPeerCloudAccount,
		"peer_vpc": testLegacyPeerVpc, "peer_region": testPeerRegion, "user_peer_network_cidrs": []tftypes.Value{},
	}
	config := tfsdk.Config{Raw: attachmentStateValue(t, values)}
	values["id"] = testLegacyFourPartID + "/" + testPeerRegion
	values["state"] = "PENDING_PEER"
	values["peering_connection_id"] = "pcx-existing"
	values["state_info"] = map[string]tftypes.Value{
		"aws_vpc_peering_connection_id": tftypes.NewValue(tftypes.String, "pcx-existing"),
		"message":                       tftypes.NewValue(tftypes.String, "waiting for acceptance"),
	}
	// The previous Framework implementation had vpc_id but no separate project
	// and project_vpc_id. With refresh disabled, the proposed plan equals that state.
	state := tfsdk.State{Raw: attachmentStateValue(t, values), Schema: sch}
	plan := tfsdk.Plan{Raw: state.Raw, Schema: sch}
	client := avngen.NewMockClient(t)
	r := adapter.NewResource(ResourceOptions)
	r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: migrationProviderData{client}}, &resource.ConfigureResponse{})
	planned := resource.ModifyPlanResponse{Plan: plan}
	r.(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), resource.ModifyPlanRequest{
		State: state, Plan: plan, Config: config,
	}, &planned)
	require.False(t, planned.Diagnostics.HasError(), "%v", planned.Diagnostics)
	require.Empty(t, planned.RequiresReplace)
	require.False(t, state.Raw.Equal(planned.Plan.Raw), "migration should populate the separate VPC fields")
	var plannedAttributes map[string]tftypes.Value
	require.NoError(t, planned.Plan.Raw.As(&plannedAttributes))
	require.False(t, plannedAttributes["peering_connection_id"].IsKnown(), "the cloud ID may disappear after rejection")

	// The resource also supports ordinary AWS VPC peerings, which the peer can reject.
	connection := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeRejectedByPeer, nil, map[string]any{"message": "rejected"})
	connection.PeerVpc = testLegacyPeerVpc
	connection.VpcPeeringConnectionType = vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection
	client.EXPECT().VpcGet(mock.Anything, testProject, testProjectVpcID).Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{connection},
	}, nil).Twice()
	updated := resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	r.Update(t.Context(), resource.UpdateRequest{State: state, Plan: planned.Plan, Config: config}, &updated)
	client.AssertNumberOfCalls(t, "VpcPeeringConnectionUpdate", 0)
	require.False(t, updated.Diagnostics.HasError(), "a state-only migration must not require an active connection: %v", updated.Diagnostics)
	var project, projectVPCID, currentState string
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("project"), &project).HasError())
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("project_vpc_id"), &projectVPCID).HasError())
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("state"), &currentState).HasError())
	require.Equal(t, testProject, project)
	require.Equal(t, testProjectVpcID, projectVPCID)
	require.Equal(t, "REJECTED_BY_PEER", currentState)
	var appliedAttributes map[string]tftypes.Value
	require.NoError(t, updated.State.Raw.As(&appliedAttributes))
	require.True(t, appliedAttributes["peering_connection_id"].IsNull())
	for name, value := range plannedAttributes {
		if value.IsKnown() {
			require.True(t, value.Equal(appliedAttributes[name]), "known planned attribute %s changed during apply", name)
		}
	}
}

func TestMigrationWithoutRefreshUnconfiguredCIDRs(t *testing.T) {
	sch := resourceSchema(t.Context())
	values := map[string]any{
		"vpc_id": testProject + "/" + testProjectVpcID, "peer_cloud_account": testPeerCloudAccount,
		"peer_vpc": testPeerVpc, "peer_region": testPeerRegion,
	}
	config := tfsdk.Config{Raw: attachmentStateValue(t, values)}
	values["id"] = testFivePartID
	values["state"] = "ACTIVE"
	values["user_peer_network_cidrs"] = []tftypes.Value{tftypes.NewValue(tftypes.String, "10.10.0.0/24")}
	state := tfsdk.State{Raw: attachmentStateValue(t, values), Schema: sch}
	plan := tfsdk.Plan{Raw: state.Raw, Schema: sch}
	client := avngen.NewMockClient(t)
	r := adapter.NewResource(ResourceOptions)
	r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: migrationProviderData{client}}, &resource.ConfigureResponse{})
	planned := resource.ModifyPlanResponse{Plan: plan}
	r.(resource.ResourceWithModifyPlan).ModifyPlan(t.Context(), resource.ModifyPlanRequest{
		State: state, Plan: plan, Config: config,
	}, &planned)
	require.False(t, planned.Diagnostics.HasError(), "%v", planned.Diagnostics)
	require.Empty(t, planned.RequiresReplace)
	require.False(t, state.Raw.Equal(planned.Plan.Raw), "migration should populate the separate VPC fields")

	// The CIDRs changed outside Terraform after the last refresh. No CIDRs are configured.
	client.EXPECT().VpcGet(mock.Anything, testProject, testProjectVpcID).Return(&vpc.VpcGetOut{
		PeeringConnections: []vpc.PeeringConnectionOut{
			newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeActive, []string{"10.20.0.0/24"}, nil),
		},
	}, nil)
	client.EXPECT().VpcPeeringConnectionUpdate(mock.Anything, testProject, testProjectVpcID, cidrUpdateMatcher(
		new(testPeerRegion), []string{"10.10.0.0/24"}, []string{"10.20.0.0/24"},
	)).Return(&vpc.VpcPeeringConnectionUpdateOut{}, nil).Maybe()
	updated := resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
	r.Update(t.Context(), resource.UpdateRequest{State: state, Plan: planned.Plan, Config: config}, &updated)
	client.AssertNumberOfCalls(t, "VpcPeeringConnectionUpdate", 0)
	require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
	var cidrs []string
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("user_peer_network_cidrs"), &cidrs).HasError())
	require.Equal(t, []string{"10.20.0.0/24"}, cidrs)

	var plannedAttributes, appliedAttributes map[string]tftypes.Value
	require.NoError(t, planned.Plan.Raw.As(&plannedAttributes))
	require.NoError(t, updated.State.Raw.As(&appliedAttributes))
	plannedCIDRs := plannedAttributes["user_peer_network_cidrs"]
	require.True(t, !plannedCIDRs.IsKnown() || plannedCIDRs.Equal(appliedAttributes["user_peer_network_cidrs"]),
		"refreshing unconfigured CIDRs must agree with the plan")
}

func TestUpdateDoesNotWaitForReadinessAfterCIDRWrite(t *testing.T) {
	check := func(t *testing.T, priorCIDR string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		sch := resourceSchema(t.Context())
		values := map[string]any{
			"id": testFivePartID, "vpc_id": testProject + "/" + testProjectVpcID,
			"project": testProject, "project_vpc_id": testProjectVpcID,
			"peer_cloud_account": testPeerCloudAccount, "peer_vpc": testPeerVpc,
			"peer_region": testPeerRegion, "state": "APPROVED",
			"user_peer_network_cidrs": []tftypes.Value{tftypes.NewValue(tftypes.String, priorCIDR)},
		}
		state := tfsdk.State{Raw: attachmentStateValue(t, values), Schema: sch}
		values["user_peer_network_cidrs"] = []tftypes.Value{tftypes.NewValue(tftypes.String, "10.20.0.0/24")}
		plan := tfsdk.Plan{Raw: attachmentStateValue(t, values), Schema: sch}
		client := avngen.NewMockClient(t)
		r := adapter.NewResource(ResourceOptions)
		r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: migrationProviderData{client}}, &resource.ConfigureResponse{})
		client.EXPECT().VpcGet(mock.Anything, testProject, testProjectVpcID).Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeApproved, []string{"10.10.0.0/24"}, nil)},
		}, nil).Once()
		client.EXPECT().VpcPeeringConnectionUpdate(mock.Anything, testProject, testProjectVpcID, cidrUpdateMatcher(
			new(testPeerRegion), []string{"10.20.0.0/24"}, []string{"10.10.0.0/24"},
		)).Return(&vpc.VpcPeeringConnectionUpdateOut{}, nil).Once()
		// PUT persists the CIDRs without changing the attachment's state.
		client.EXPECT().VpcGet(mock.Anything, testProject, testProjectVpcID).Return(&vpc.VpcGetOut{
			PeeringConnections: []vpc.PeeringConnectionOut{newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeApproved, []string{"10.20.0.0/24"}, nil)},
		}, nil).Once()
		updated := resource.UpdateResponse{State: tfsdk.State{Schema: sch}}
		r.Update(ctx, resource.UpdateRequest{State: state, Plan: plan, Config: tfsdk.Config{Raw: plan.Raw}}, &updated)
		require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
		var cidrs []string
		require.False(t, updated.State.GetAttribute(t.Context(), path.Root("user_peer_network_cidrs"), &cidrs).HasError())
		require.Equal(t, []string{"10.20.0.0/24"}, cidrs)
	}
	t.Run("configured CIDRs changed", func(t *testing.T) { check(t, "10.10.0.0/24") })
	t.Run("unchanged CIDRs drifted in the API", func(t *testing.T) { check(t, "10.20.0.0/24") })
}

func TestUpdateDoesNotWaitForReadinessAfterReadReconcilesCIDRs(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	sch := resourceSchema(t.Context())
	state := tfsdk.State{Schema: sch, Raw: attachmentStateValue(t, map[string]any{
		"id": testFivePartID, "vpc_id": testProject + "/" + testProjectVpcID,
		"peer_cloud_account": testPeerCloudAccount, "peer_vpc": testPeerVpc, "peer_region": testPeerRegion,
		"state":                   "APPROVED",
		"user_peer_network_cidrs": []tftypes.Value{tftypes.NewValue(tftypes.String, "10.20.0.0/24")},
	})}
	client := avngen.NewMockClient(t)
	// Update itself has nothing to write. CIDRs drift before its subsequent Read.
	expectVpcGet(t, client, newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeApproved, []string{"10.20.0.0/24"}, nil))
	expectVpcGet(t, client, newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypeApproved, []string{"10.10.0.0/24"}, nil))
	client.EXPECT().VpcPeeringConnectionUpdate(mock.Anything, testProject, testProjectVpcID, cidrUpdateMatcher(
		new(testPeerRegion), []string{"10.20.0.0/24"}, []string{"10.10.0.0/24"},
	)).Return(&vpc.VpcPeeringConnectionUpdateOut{}, nil).Once()
	r := adapter.NewResource(ResourceOptions)
	r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: migrationProviderData{client}}, &resource.ConfigureResponse{})
	updated := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{
		State: state, Plan: tfsdk.Plan{Raw: state.Raw, Schema: sch}, Config: tfsdk.Config{Raw: state.Raw},
	}, &updated)
	require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
	var cidrs []string
	require.False(t, updated.State.GetAttribute(t.Context(), path.Root("user_peer_network_cidrs"), &cidrs).HasError())
	require.Equal(t, []string{"10.20.0.0/24"}, cidrs)
}

func TestReadThenDeleteCrossRegionImport(t *testing.T) {
	client := avngen.NewMockClient(t)
	connection := newConnection(testPeerRegion, vpc.VpcPeeringConnectionStateTypePendingPeer, nil, nil)
	connection.PeerVpc = testLegacyPeerVpc
	connection.VpcPeeringConnectionType = vpc.VpcPeeringConnectionTypeAWSVpcPeeringConnection
	connections := []vpc.PeeringConnectionOut{connection}
	client.EXPECT().VpcGet(mock.Anything, testProject, testProjectVpcID).
		RunAndReturn(func(context.Context, string, string) (*vpc.VpcGetOut, error) {
			return &vpc.VpcGetOut{CloudName: "aws-eu-west-1", PeeringConnections: connections}, nil
		})
	// A regionless DELETE targets eu-west-1, where this connection does not exist.
	client.EXPECT().VpcPeeringConnectionDelete(mock.Anything, testProject, testProjectVpcID, testPeerCloudAccount, testLegacyPeerVpc).
		Return(nil, avngen.Error{Status: 404}).Maybe()
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(mock.Anything, testProject, testProjectVpcID, testPeerCloudAccount, testLegacyPeerVpc, testPeerRegion).
		RunAndReturn(func(context.Context, string, string, string, string, string) (*vpc.VpcPeeringConnectionWithRegionDeleteOut, error) {
			connections = nil
			return &vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil
		}).Once()
	r := adapter.NewResource(ResourceOptions)
	r.(resource.ResourceWithConfigure).Configure(t.Context(), resource.ConfigureRequest{ProviderData: migrationProviderData{client}}, &resource.ConfigureResponse{})
	state := tfsdk.State{Schema: resourceSchema(t.Context()), Raw: attachmentStateValue(t, map[string]any{"id": testLegacyFourPartID})}
	read := resource.ReadResponse{State: state}
	r.Read(t.Context(), resource.ReadRequest{State: state}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	var id string
	require.False(t, read.State.GetAttribute(t.Context(), path.Root("id"), &id).HasError())
	require.Equal(t, testLegacyFourPartID, id)
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	deleted := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &deleted)
	require.False(t, deleted.Diagnostics.HasError(), "%v", deleted.Diagnostics)
	require.Empty(t, connections, "successful Delete must remove the connection while keeping its parent VPC")
}

func parseID(t *testing.T, value string) *pluginvpc.ProjectPeeringID {
	t.Helper()

	id, err := pluginvpc.ParseProjectPeeringID(value)
	require.NoError(t, err)
	return id
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

func attachmentStateValue(t *testing.T, values map[string]any) tftypes.Value {
	t.Helper()
	typ := resourceSchema(t.Context()).Type().TerraformType(t.Context()).(tftypes.Object)
	attributes := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attributeType := range typ.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, values[name])
	}
	return tftypes.NewValue(typ, attributes)
}

type migrationProviderData struct{ client avngen.Client }

func (migrationProviderData) GetClient() *aiven.Client      { return nil }
func (p migrationProviderData) GetGenClient() avngen.Client { return p.client }
