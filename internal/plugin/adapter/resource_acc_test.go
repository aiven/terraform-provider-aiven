package adapter_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	orgvpc "github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/organizationvpc"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/projectvpc"
)

// Exercise Terraform's identity consistency checks without cloud credentials.
func TestResourceIdentity(t *testing.T) {
	const address = "aiven_identity_test.example"
	const config = `resource "aiven_identity_test" "example" { name = "updated" }`
	var mu sync.Mutex
	var name string
	var deleted bool
	read := func(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
		mu.Lock()
		defer mu.Unlock()
		if deleted {
			return adapter.ErrNotFound
		}
		assert.Equal(t, "remote-id", d.ID())
		// The API resolves an optional identity field omitted during import.
		return d.Flatten(map[string]any{"id": "remote-id", "region": "eu-west-1", "name": name})
	}
	write := func(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
		mu.Lock()
		defer mu.Unlock()
		name, deleted = d.Get("name").(string), false
		return d.Flatten(map[string]any{"id": "remote-id", "region": "eu-west-1", "name": name})
	}
	opts := adapter.ResourceOptions{
		TypeName: "aiven_identity_test",
		Schema: func(ctx context.Context) schema.Schema {
			return schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":     schema.StringAttribute{Computed: true},
					"region": schema.StringAttribute{Computed: true},
					"name":   schema.StringAttribute{Required: true},
				},
				Blocks: map[string]schema.Block{"timeouts": timeouts.BlockAll(ctx)},
			}
		},
		SchemaInternal: &adapter.Schema{Type: adapter.SchemaTypeObject, Properties: map[string]*adapter.Schema{
			"id":     {Type: adapter.SchemaTypeString, Computed: true},
			"region": {Type: adapter.SchemaTypeString, Computed: true},
			"name":   {Type: adapter.SchemaTypeString},
			"timeouts": {Type: adapter.SchemaTypeObject, Properties: map[string]*adapter.Schema{
				"create": {Type: adapter.SchemaTypeString}, "read": {Type: adapter.SchemaTypeString},
				"update": {Type: adapter.SchemaTypeString}, "delete": {Type: adapter.SchemaTypeString},
			}},
		}},
		IDFields:               []string{"id"},
		IdentityFields:         []string{"id"},
		IdentityFieldsOptional: []string{"region"},
		RemoveMissing:          true,
		Create:                 write,
		Read:                   read,
		Update:                 write,
		Delete: func(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
			mu.Lock()
			defer mu.Unlock()
			assert.Equal(t, "remote-id", d.ID())
			assert.Equal(t, "eu-west-1", d.Get("region"))
			deleted = true
			return nil
		},
	}
	var identityImport acc.IdentityImportVersionCheck
	identity := []statecheck.StateCheck{statecheck.ExpectIdentity(address, map[string]knownvalue.Check{
		"id": knownvalue.StringExact("remote-id"), "region": knownvalue.StringExact("eu-west-1"),
	})}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&identityTestProvider{
				AivenProvider: plugin.AivenProvider{GenClient: avngen.NewMockClient(t)},
				options:       []adapter.ResourceOptions{opts},
			}),
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
		Steps: []resource.TestStep{
			{
				Config: `resource "aiven_identity_test" "example" { name = "initial" }`,
				Check:  resource.TestCheckResourceAttr(address, "region", "eu-west-1"),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
				}},
			},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
			{
				ResourceName: address, ImportState: true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				SkipFunc:        identityImport.Skip,
			},
			{
				Config: `removed {
  from = aiven_identity_test.example
  lifecycle { destroy = false }
}`,
				SkipFunc: identityImport.Skip,
			},
			{
				Config: config + `
import {
  to = aiven_identity_test.example
  identity = { region = "eu-west-1" }
}`,
				SkipFunc:    identityImport.Skip,
				ExpectError: regexp.MustCompile(`attribute "id" is required`),
			},
			{
				// Import without region must return the same complete identity as Create.
				Config: config + `
import {
  to = aiven_identity_test.example
  identity = { id = "remote-id" }
}`,
				SkipFunc:          identityImport.Skip,
				ConfigStateChecks: identity,
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				}},
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					deleted = true
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreate),
				}},
			},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	require.True(t, deleted)
}

type identityTestProvider struct {
	plugin.AivenProvider
	options []adapter.ResourceOptions
}

func (p *identityTestProvider) Resources(context.Context) []func() frameworkresource.Resource {
	result := make([]func() frameworkresource.Resource, len(p.options))
	for i, opts := range p.options {
		result[i] = adapter.NewLazyResource(opts)
	}
	return result
}

const (
	identityVPCID          = "12345678-1234-4234-8234-123456789012"
	identityProjectAddress = "aiven_project_vpc.example"
	identityOrgAddress     = "aiven_organization_vpc.example"
	identityVPCConfig      = `
resource "aiven_project_vpc" "example" {
  project      = "project"
  cloud_name   = "google-europe-west2"
  network_cidr = "192.168.0.0/24"
}
resource "aiven_organization_vpc" "example" {
  organization_id = "org1234567890"
  cloud_name      = "google-europe-west2"
  network_cidr    = "192.168.1.0/24"
}`
)

func TestResourceIdentityVPCs(t *testing.T) {
	var identityEnabled atomic.Bool
	var identityImport acc.IdentityImportVersionCheck
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": identityVPCProviderFactory(t, &identityEnabled),
		},
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
		Steps: []resource.TestStep{
			{Config: identityVPCConfig},
			{
				// Adding identity to existing state must not change the plan or string IDs.
				PreConfig: func() { identityEnabled.Store(true) },
				Config:    identityVPCConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(identityProjectAddress, "id", "project/"+identityVPCID),
					resource.TestCheckResourceAttr(identityOrgAddress, "id", "org1234567890/"+identityVPCID),
				),
			},
			{ResourceName: identityProjectAddress, ImportState: true, ImportStateVerify: true},
			{ResourceName: identityOrgAddress, ImportState: true, ImportStateVerify: true},
			{
				ResourceName: identityProjectAddress, ImportState: true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
			},
			{
				ResourceName: identityOrgAddress, ImportState: true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
			},
		},
	})
}

// Unlike separate runs of TestResourceIdentityVPCs, this test changes Terraform
// versions against the same on-disk state. Only the Aiven API is mocked.
// Set TF_ACC_TERRAFORM_111_PATH and TF_ACC_TERRAFORM_112_PATH to reuse local
// binaries; otherwise the pinned Terraform releases are downloaded.
func TestAccResourceIdentityTerraformUpgrade(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) != "1" {
		t.Skip("set TF_ACC=1 to run the test with Terraform 1.11 and 1.12")
	}
	ctx := t.Context()
	dir := t.TempDir()
	newCLI := func(pinnedVersion, env string) *tfexec.Terraform {
		t.Helper()
		binary := os.Getenv(env)
		if binary == "" {
			installer := releases.ExactVersion{
				Product: product.Terraform, Version: version.Must(version.NewVersion(pinnedVersion)),
				InstallDir: t.TempDir(),
			}
			var err error
			binary, err = installer.Install(ctx)
			require.NoError(t, err)
		}
		tf, err := tfexec.NewTerraform(dir, binary)
		require.NoError(t, err)
		require.NoError(t, tf.SetEnv(map[string]string{"TF_CLI_CONFIG_FILE": os.DevNull, "CHECKPOINT_DISABLE": "1"}))
		actual, _, err := tf.Version(ctx, true)
		require.NoError(t, err)
		minor := pinnedVersion[:strings.LastIndex(pinnedVersion, ".")+1]
		require.True(t, strings.HasPrefix(actual.String(), minor), "%s must point to Terraform %sx", env, minor)
		t.Logf("Using Terraform %s: %s", actual, binary)
		return tf
	}
	oldTF := newCLI("1.11.4", "TF_ACC_TERRAFORM_111_PATH")
	newTF := newCLI("1.12.2", "TF_ACC_TERRAFORM_112_PATH")
	var identityEnabled atomic.Bool
	factory := identityVPCProviderFactory(t, &identityEnabled)
	startProvider := func() (*tfexec.ReattachOption, func()) {
		t.Helper()
		provider, err := factory()
		require.NoError(t, err)
		serverCtx, cancel := context.WithCancel(ctx)
		ready := make(chan *goplugin.ReattachConfig, 1)
		closed := make(chan struct{})
		serveErr := make(chan error, 1)
		go func() {
			serveErr <- tf6server.Serve("registry.terraform.io/aiven/aiven", func() tfprotov6.ProviderServer { return provider },
				tf6server.WithDebug(serverCtx, ready, closed), tf6server.WithLoggingSink(t),
				tf6server.WithGoPluginLogger(hclog.NewNullLogger()))
		}()
		stop := func() {
			cancel()
			select {
			case <-closed:
			case <-time.After(5 * time.Second):
				t.Error("provider did not stop")
			}
		}
		t.Cleanup(stop)
		select {
		case r := <-ready:
			return tfexec.Reattach(tfexec.ReattachInfo{"registry.terraform.io/aiven/aiven": {
				Protocol: string(r.Protocol), ProtocolVersion: r.ProtocolVersion, Pid: r.Pid, Test: r.Test,
				Addr: tfexec.ReattachConfigAddr{Network: r.Addr.Network(), String: r.Addr.String()},
			}}), stop
		case err := <-serveErr:
			t.Fatalf("starting provider: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("provider did not start")
		}
		return nil, stop
	}
	const requiredProvider = `
terraform {
  required_providers {
    aiven = { source = "aiven/aiven" }
  }
}
`
	writeConfig := func(extra string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(requiredProvider+identityVPCConfig+extra), 0o600))
	}
	writeConfig("")
	reattach, stop := startProvider()
	require.NoError(t, oldTF.Init(ctx, reattach))
	require.NoError(t, oldTF.Apply(ctx, reattach))
	state, err := oldTF.Show(ctx, reattach)
	require.NoError(t, err)
	require.Len(t, state.Values.RootModule.Resources, 2)
	for _, r := range state.Values.RootModule.Resources {
		require.Empty(t, r.IdentityValues)
	}
	stop()
	identityEnabled.Store(true)
	reattach, _ = startProvider()

	// Upgrading the provider on 1.11 and then Terraform itself must also work
	// without refresh: identity can be populated on a subsequent refresh.
	for _, tf := range []*tfexec.Terraform{oldTF, newTF} {
		hasChanges, err := tf.Plan(ctx, reattach, tfexec.Refresh(false))
		require.NoError(t, err)
		require.False(t, hasChanges)
		require.NoError(t, tf.Apply(ctx, reattach, tfexec.Refresh(false)))
	}
	checkIdentity := func() {
		t.Helper()
		state, err := newTF.Show(ctx, reattach)
		require.NoError(t, err)
		require.Len(t, state.Values.RootModule.Resources, 2)
		for _, r := range state.Values.RootModule.Resources {
			switch r.Address {
			case identityProjectAddress:
				require.Equal(t, "project/"+identityVPCID, r.AttributeValues["id"])
				require.Equal(t, map[string]any{"project": "project", "project_vpc_id": identityVPCID}, r.IdentityValues)
			case identityOrgAddress:
				require.Equal(t, "org1234567890/"+identityVPCID, r.AttributeValues["id"])
				require.Equal(t, map[string]any{"organization_id": "org1234567890", "organization_vpc_id": identityVPCID}, r.IdentityValues)
			default:
				t.Fatalf("unexpected resource %s", r.Address)
			}
		}
	}
	hasChanges, err := newTF.Plan(ctx, reattach)
	require.NoError(t, err)
	require.False(t, hasChanges)
	require.NoError(t, newTF.Apply(ctx, reattach))
	checkIdentity()

	// Import into real state, then plan and destroy the imported instances.
	// TestStep.ImportBlockWithResourceIdentity only verifies the import plan.
	require.NoError(t, newTF.StateRm(ctx, identityProjectAddress))
	require.NoError(t, newTF.StateRm(ctx, identityOrgAddress))
	writeConfig(`
import {
  to = aiven_project_vpc.example
  identity = {
    project        = "project"
    project_vpc_id = "12345678-1234-4234-8234-123456789012"
  }
}
import {
  to = aiven_organization_vpc.example
  identity = {
    organization_id     = "org1234567890"
    organization_vpc_id = "12345678-1234-4234-8234-123456789012"
  }
}`)
	require.NoError(t, newTF.Apply(ctx, reattach))
	checkIdentity()
	writeConfig("")
	// A rollback to 1.11 must remain usable after 1.12 has written identity.
	// Returning to 1.12 must restore identity from the unchanged resource fields.
	hasChanges, err = oldTF.Plan(ctx, reattach)
	require.NoError(t, err)
	require.False(t, hasChanges)
	require.NoError(t, oldTF.Apply(ctx, reattach))
	require.NoError(t, newTF.Apply(ctx, reattach))
	checkIdentity()
	hasChanges, err = newTF.Plan(ctx, reattach)
	require.NoError(t, err)
	require.False(t, hasChanges)
	require.NoError(t, newTF.Destroy(ctx, reattach))
}

func identityVPCProviderFactory(t *testing.T, identityEnabled *atomic.Bool) func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	client := avngen.NewMockClient(t)
	var projectDeleted, orgDeleted atomic.Bool
	client.EXPECT().VpcCreate(mock.Anything, "project", mock.Anything).
		Return(&vpc.VpcCreateOut{ProjectVpcId: identityVPCID}, nil).Once()
	client.EXPECT().VpcGet(mock.Anything, "project", identityVPCID).
		RunAndReturn(func(context.Context, string, string) (*vpc.VpcGetOut, error) {
			state := vpc.VpcStateTypeActive
			if projectDeleted.Load() {
				state = vpc.VpcStateTypeDeleted
			}
			return &vpc.VpcGetOut{
				ProjectVpcId: identityVPCID, State: state,
				CloudName: "google-europe-west2", NetworkCidr: "192.168.0.0/24",
			}, nil
		})
	client.EXPECT().VpcDelete(mock.Anything, "project", identityVPCID).
		RunAndReturn(func(context.Context, string, string) (*vpc.VpcDeleteOut, error) {
			projectDeleted.Store(true)
			return &vpc.VpcDeleteOut{}, nil
		}).Once()
	client.EXPECT().OrganizationVpcCreate(mock.Anything, "org1234567890", mock.Anything).
		Return(&orgvpc.OrganizationVpcCreateOut{
			OrganizationId: "org1234567890", OrganizationVpcId: identityVPCID,
			Clouds: []orgvpc.CloudOut{{CloudName: "google-europe-west2", NetworkCidr: "192.168.1.0/24"}},
		}, nil).Once()
	client.EXPECT().OrganizationVpcGet(mock.Anything, "org1234567890", identityVPCID).
		RunAndReturn(func(context.Context, string, string) (*orgvpc.OrganizationVpcGetOut, error) {
			state := orgvpc.OrganizationVpcStateTypeActive
			if orgDeleted.Load() {
				state = orgvpc.OrganizationVpcStateTypeDeleted
			}
			return &orgvpc.OrganizationVpcGetOut{
				OrganizationId: "org1234567890", OrganizationVpcId: identityVPCID, State: state,
				Clouds: []orgvpc.CloudOut{{CloudName: "google-europe-west2", NetworkCidr: "192.168.1.0/24"}},
			}, nil
		})
	client.EXPECT().OrganizationVpcDelete(mock.Anything, "org1234567890", identityVPCID).
		RunAndReturn(func(context.Context, string, string) (*orgvpc.OrganizationVpcDeleteOut, error) {
			orgDeleted.Store(true)
			return &orgvpc.OrganizationVpcDeleteOut{}, nil
		}).Once()
	t.Cleanup(func() {
		if !t.Failed() {
			require.True(t, projectDeleted.Load())
			require.True(t, orgDeleted.Load())
		}
	})
	return func() (tfprotov6.ProviderServer, error) {
		opts := []adapter.ResourceOptions{projectvpc.ResourceOptions, organizationvpc.ResourceOptions}
		if !identityEnabled.Load() {
			for i := range opts {
				opts[i].IdentityFields = nil
			}
		}
		return providerserver.NewProtocol6WithError(&identityTestProvider{
			AivenProvider: plugin.AivenProvider{GenClient: client}, options: opts,
		})()
	}
}
