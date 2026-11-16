package servicecompat_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"sync/atomic"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// SDKv2 removed a service deleted outside Terraform from the state, so the next plan creates it.
func TestDeletedService(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	f := grafanaFixture()
	client := newClient(t, f, func(c *avngen.MockClient) {
		c.EXPECT().
			ServiceGet(mock.Anything, testProject, f.state["service_name"], mock.Anything).
			Return(nil, avngen.Error{Status: 404, Message: "Service not found"}).
			Once()
	})
	server, schemas := newServer(t, client)

	raw, err := json.Marshal(f.state)
	require.NoError(t, err)
	up, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: f.typeName,
		Version:  f.version,
		RawState: &tfprotov6.RawState{JSON: raw},
	})
	require.NoError(t, err)
	requireNoErrors(t, up.Diagnostics)

	read, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: f.typeName, CurrentState: up.UpgradedState})
	require.NoError(t, err)
	requireNoErrors(t, read.Diagnostics)
	assert.True(t, unmarshal(t, read.NewState, schemas[f.typeName].ValueType()).IsNull())
}

// A create that fails after the service is created keeps it in the state, or the next apply creates it again.
func TestCreateFailure(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		fail func(c *avngen.MockClient, created *atomic.Bool)
	}{
		{
			// The create response flatten reads the plan
			name: "flatten",
			fail: func(c *avngen.MockClient, created *atomic.Bool) {
				c.EXPECT().ProjectServicePlanSpecsGet(mock.Anything, testProject, mock.Anything, mock.Anything).
					RunAndReturn(func(context.Context, string, string, string) (*project.ProjectServicePlanSpecsGetOut, error) {
						if created.Load() {
							return nil, errors.New("plan specs failed")
						}
						return grafanaFixture().specs, nil
					}).
					Maybe()
			},
		},
		{
			name: "wait",
			fail: func(c *avngen.MockClient, _ *atomic.Bool) {
				c.EXPECT().ServiceGet(mock.Anything, testProject, "my-grafana").
					Return(nil, avngen.Error{Status: 500, Message: "wait failed"}).
					Maybe()
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			f := grafanaFixture()
			var created atomic.Bool
			client := newClient(t, f, func(c *avngen.MockClient) {
				c.EXPECT().ServiceCreate(mock.Anything, testProject, mock.Anything).
					RunAndReturn(func(context.Context, string, *service.ServiceCreateIn) (*service.ServiceCreateOut, error) {
						created.Store(true)
						out := new(service.ServiceCreateOut)
						b, err := json.Marshal(f.api)
						if err != nil {
							return nil, err
						}
						return out, json.Unmarshal(b, out)
					}).
					Once()
				tc.fail(c, &created)
			})
			server, schemas := newServer(t, client)
			sch := schemas[f.typeName]
			typ := sch.ValueType()

			config := dynamicValue(t, typ, valueForBlock(sch.Block, f.config))
			prior := dynamicValue(t, typ, tftypes.NewValue(typ, nil))
			plan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName:         f.typeName,
				Config:           &config,
				PriorState:       &prior,
				ProposedNewState: &config,
			})
			require.NoError(t, err)
			requireNoErrors(t, plan.Diagnostics)

			rsp, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{
				TypeName:        f.typeName,
				PriorState:      &prior,
				PlannedState:    plan.PlannedState,
				Config:          &config,
				PlannedPrivate:  plan.PlannedPrivate,
				PlannedIdentity: plan.PlannedIdentity,
			})
			require.NoError(t, err)
			require.True(t, created.Load())
			require.NotEmpty(t, rsp.Diagnostics)

			var attrs map[string]tftypes.Value
			require.NoError(t, unmarshal(t, rsp.NewState, typ).As(&attrs))
			assert.True(t, attrs["id"].Equal(tftypes.NewValue(tftypes.String, testProject+"/my-grafana")))
		})
	}
}

// Removing additional_disk_space shrinks the disk to the plan default, as SDKv2 did.
func TestRemovedAdditionalDiskSpace(t *testing.T) {
	t.Parallel()
	f := pgFixture()
	maps.Copy(f.state, map[string]any{"additional_disk_space": "30GiB", "disk_space_used": "110GiB"})
	f.api.DiskSpaceMb = new(112640)

	// The mock can't wait for the update, so it stops at the request
	var sent *service.ServiceUpdateIn
	client := newClient(t, f, func(c *avngen.MockClient) {
		c.EXPECT().ServiceUpdate(mock.Anything, testProject, "my-pg", mock.Anything).
			RunAndReturn(func(_ context.Context, _, _ string, in *service.ServiceUpdateIn, _ ...[2]string) (*service.ServiceUpdateOut, error) {
				sent = in
				return nil, errors.New("stop")
			}).
			Once()
	})

	p := planUpgrade(t, f, client)
	var attrs map[string]tftypes.Value
	require.NoError(t, p.planned.As(&attrs))
	assert.True(t, attrs["additional_disk_space"].Equal(tftypes.NewValue(tftypes.String, "0B")))
	for _, k := range []string{"service_host", "service_password", "service_port", "service_uri", "service_username"} {
		assert.True(t, attrs[k].IsKnown(), k)
	}

	p.apply(t)
	require.NotNil(t, sent)
	assert.Equal(t, 81920, *sent.DiskSpaceMb)
}

// Terraform marks a sensitive attribute nested in another attribute as a whole, without the attributes in it.
// SDKv2 states mark those one by one, so the marks would differ, and the first plan would update the resource.
func TestSensitiveNestedAttributes(t *testing.T) {
	t.Parallel()
	_, schemas := newServer(t, avngen.NewMockClient(t))
	for name, s := range schemas {
		assertSensitiveBlock(t, name, s.Block)
	}
}

func assertSensitiveBlock(t *testing.T, path string, block *tfprotov6.SchemaBlock) {
	for _, attr := range block.Attributes {
		if attr.NestedType != nil {
			assertSensitiveObject(t, path+"."+attr.Name, attr.NestedType)
		}
	}
	for _, nested := range block.BlockTypes {
		assertSensitiveBlock(t, path+"."+nested.TypeName, nested.Block)
	}
}

func assertSensitiveObject(t *testing.T, path string, object *tfprotov6.SchemaObject) {
	for _, attr := range object.Attributes {
		if attr.NestedType != nil {
			assert.False(t, attr.Sensitive, "%s.%s", path, attr.Name)
			assertSensitiveObject(t, path+"."+attr.Name, attr.NestedType)
		}
	}
}
