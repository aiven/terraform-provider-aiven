package byoc_test

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

const (
	providerAddress = "registry.terraform.io/aiven/aiven"
	providerConfig  = `
terraform {
  required_providers {
    aiven = { source = "aiven/aiven" }
  }
}
`
)

// Keep the same CLI working directory after failed applies, imports and destroys,
// so the next command sees exactly the state Terraform left for the user.
type azureTerraform struct {
	t        *testing.T
	cli      *tfexec.Terraform
	reattach *tfexec.ReattachOption
}

func newAzureTerraform(t *testing.T, api *cloudAPI, config string) *azureTerraform {
	t.Helper()
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	client, err := avngen.NewClient(avngen.TokenOpt("test"), avngen.HostOpt(server.URL), avngen.DoerOpt(server.Client()))
	require.NoError(t, err)

	binary := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if binary == "" {
		binary, err = exec.LookPath("terraform")
		require.NoError(t, err)
	}
	cli, err := tfexec.NewTerraform(t.TempDir(), binary)
	require.NoError(t, err)
	require.NoError(t, cli.SetEnv(map[string]string{"TF_CLI_CONFIG_FILE": os.DevNull, "CHECKPOINT_DISABLE": "1"}))

	ctx, cancel := context.WithCancel(t.Context())
	ready := make(chan *goplugin.ReattachConfig, 1)
	done := make(chan struct{})
	var serveErr error
	go func() {
		defer close(done)
		serveErr = tf6server.Serve(providerAddress, providerserver.NewProtocol6(&plugin.AivenProvider{GenClient: client}),
			tf6server.WithDebug(ctx, ready, make(chan struct{})), tf6server.WithLoggingSink(t),
			tf6server.WithGoPluginLogger(hclog.NewNullLogger()))
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
			require.NoError(t, serveErr)
		case <-time.After(10 * time.Second):
			t.Error("Terraform provider did not stop")
		}
	})
	var connection *goplugin.ReattachConfig
	select {
	case connection = <-ready:
	case <-done:
		t.Fatalf("Terraform provider stopped during startup: %v", serveErr)
	case <-time.After(10 * time.Second):
		t.Fatal("Terraform provider did not start")
	}
	tf := &azureTerraform{t: t, cli: cli, reattach: tfexec.Reattach(tfexec.ReattachInfo{providerAddress: {
		Protocol: string(connection.Protocol), ProtocolVersion: connection.ProtocolVersion, Pid: connection.Pid, Test: connection.Test,
		Addr: tfexec.ReattachConfigAddr{Network: connection.Addr.Network(), String: connection.Addr.String()},
	}})}
	tf.config(config)
	require.NoError(t, cli.Init(t.Context(), tf.reattach))
	return tf
}

func (tf *azureTerraform) config(resources string) {
	tf.t.Helper()
	require.NoError(tf.t, os.WriteFile(filepath.Join(tf.cli.WorkingDir(), "main.tf"), []byte(providerConfig+resources), 0o600))
}

func (tf *azureTerraform) apply() error {
	return tf.cli.Apply(tf.t.Context(), tf.reattach)
}

func (tf *azureTerraform) destroy() error {
	return tf.cli.Destroy(tf.t.Context(), tf.reattach)
}

func (tf *azureTerraform) importResource(address, id string) error {
	return tf.cli.Import(tf.t.Context(), address, "org1234567890/"+id, tf.reattach)
}

func (tf *azureTerraform) resource(address string) map[string]any {
	tf.t.Helper()
	state, err := tf.cli.Show(tf.t.Context(), tf.reattach)
	require.NoError(tf.t, err)
	if state.Values != nil && state.Values.RootModule != nil {
		for _, r := range state.Values.RootModule.Resources {
			if r.Address == address {
				return r.AttributeValues
			}
		}
	}
	return nil
}

func (tf *azureTerraform) requireEmptyPlan() {
	tf.t.Helper()
	changed, err := tf.cli.Plan(tf.t.Context(), tf.reattach)
	require.NoError(tf.t, err)
	require.False(tf.t, changed, "expected an empty Terraform plan")
}

func (tf *azureTerraform) plannedActions(address string) tfjson.Actions {
	tf.t.Helper()
	path := filepath.Join(tf.cli.WorkingDir(), "test.tfplan")
	_, err := tf.cli.Plan(tf.t.Context(), tf.reattach, tfexec.Out(path))
	require.NoError(tf.t, err)
	plan, err := tf.cli.ShowPlanFile(tf.t.Context(), path, tf.reattach)
	require.NoError(tf.t, err)
	for _, change := range plan.ResourceChanges {
		if change.Address == address {
			return change.Change.Actions
		}
	}
	tf.t.Fatalf("missing %s in plan", address)
	return nil
}
