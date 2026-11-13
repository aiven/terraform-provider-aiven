package gcpprivatelinkconnectionapproval_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/privatelink"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	approvalAddress = "aiven_gcp_privatelink_connection_approval.test"
	firstPSC        = "50079288774361602"
	secondPSC       = "50079288774361603"
)

func TestAccAivenGCPPrivatelinkConnectionApproval_basic(t *testing.T) {
	gcpProject := acc.RequireEnvVars(t, "GCP_PRIVATE_LINK_TEST", "GOOGLE_PROJECT")["GOOGLE_PROJECT"]
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	const approvalName = "aiven_gcp_privatelink_connection_approval.endpoint"
	infrastructure := acc.GCPPrivatelinkFixture(rName) + acc.GCPPrivatelinkEndpointFixture(rName, gcpProject)
	config := infrastructure + testAccGCPPrivatelinkConnectionApprovalResource()
	approvalChecks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(approvalName, "privatelink_connection_id"),
		resource.TestCheckResourceAttrPair(approvalName, "psc_connection_id", "google_compute_forwarding_rule.endpoint", "psc_connection_id"),
		resource.TestCheckResourceAttrPair(approvalName, "user_ip_address", "google_compute_address.endpoint", "address"),
		resource.TestMatchResourceAttr(approvalName, "state", regexp.MustCompile("^(user-approved|active)$")),
	)
	var connectionID string
	var identityImport acc.IdentityImportVersionCheck
	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"google": {Source: "hashicorp/google", VersionConstraint: "=6.15.0"},
		},
		CheckDestroy: acc.TestAccCheckGCPPrivatelinkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					approvalChecks,
					func(state *terraform.State) error {
						// Approval completes at user-approved. Wait for active, then let
						// the next step refresh Terraform state before verifying imports.
						// connected has no approved IP and isn't an intermediate approval state.
						attrs := state.RootModule().Resources[approvalName].Primary.Attributes
						connectionID = attrs["privatelink_connection_id"]
						client := acc.GetTestAivenClient().GCPPrivatelink
						ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
						defer cancel()
						var nextRefresh time.Time
						wait := &retry.StateChangeConf{
							Pending: []string{"pending-user-approval", "user-approved"},
							Target:  []string{"active"},
							Refresh: func() (any, string, error) {
								if !time.Now().Before(nextRefresh) {
									if err := client.Refresh(ctx, attrs["project"], attrs["service_name"]); err != nil {
										return nil, "", err
									}
									nextRefresh = time.Now().Add(time.Minute)
								}
								connection, err := client.ConnectionGet(ctx, attrs["project"], attrs["service_name"], &connectionID)
								if err != nil {
									return nil, "", err
								}
								return connection, connection.State, nil
							},
							Timeout:      10 * time.Minute,
							PollInterval: 5 * time.Second,
						}
						if _, err := wait.WaitForStateContext(ctx); err != nil {
							return fmt.Errorf("wait for active GCP PrivateLink connection %q: %w", connectionID, err)
						}
						return nil
					},
				),
			},
			{
				Config:           config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check: resource.ComposeTestCheckFunc(
					approvalChecks,
					resource.TestCheckResourceAttr(approvalName, "state", "active"),
				),
			},
			{ResourceName: approvalName, ImportState: true, ImportStateVerify: true},
			{
				ResourceName: approvalName, ImportState: true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
			},
			{
				ResourceName: approvalName, ImportState: true, ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					attrs := state.RootModule().Resources[approvalName].Primary.Attributes
					return schemautil.BuildResourceID(attrs["project"], attrs["service_name"], attrs["psc_connection_id"]), nil
				},
			},
			{
				// Approval deletion is state-only; keep the consumer endpoint to verify it.
				Config: infrastructure,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aiven_gcp_privatelink.foo", "state", "active"),
					resource.TestCheckResourceAttrSet("google_compute_forwarding_rule.endpoint", "psc_connection_id"),
					func(state *terraform.State) error {
						if _, ok := state.RootModule().Resources[approvalName]; ok {
							return fmt.Errorf("approval %q remains in Terraform state after deletion", approvalName)
						}
						attrs := state.RootModule().Resources["aiven_gcp_privatelink.foo"].Primary.Attributes
						connection, err := acc.GetTestAivenClient().GCPPrivatelink.ConnectionGet(t.Context(), attrs["project"], attrs["service_name"], &connectionID)
						if err != nil {
							return fmt.Errorf("get GCP PrivateLink connection after approval deletion: %w", err)
						}
						if connection.State != "active" {
							return fmt.Errorf("expected connection to remain active after approval deletion, got %q", connection.State)
						}
						ip := state.RootModule().Resources["google_compute_address.endpoint"].Primary.Attributes["address"]
						if connection.UserIPAddress != ip {
							return fmt.Errorf("expected connection IP %q after approval deletion, got %q", ip, connection.UserIPAddress)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAivenGCPPrivatelinkConnectionApproval_backwardCompat(t *testing.T) {
	gcpProject := acc.RequireEnvVars(t, "GCP_PRIVATE_LINK_TEST", "GOOGLE_PROJECT")["GOOGLE_PROJECT"]
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	const approvalName = "aiven_gcp_privatelink_connection_approval.endpoint"
	config := acc.GCPPrivatelinkFixture(rName) + acc.GCPPrivatelinkEndpointFixture(rName, gcpProject) + testAccGCPPrivatelinkConnectionApprovalResource()
	steps := acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
		TFConfig: config, OldProviderVersion: "4.65.1",
		Checks: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(approvalName, "id", acc.ProjectName()+"/test-acc-sr-"+rName),
			resource.TestCheckResourceAttrSet(approvalName, "privatelink_connection_id"),
			resource.TestCheckResourceAttrPair(approvalName, "psc_connection_id", "google_compute_forwarding_rule.endpoint", "psc_connection_id"),
			resource.TestCheckResourceAttrPair(approvalName, "user_ip_address", "google_compute_address.endpoint", "address"),
			resource.TestMatchResourceAttr(approvalName, "state", regexp.MustCompile("^(user-approved|active)$")),
		),
	})
	googleProvider := resource.ExternalProvider{Source: "hashicorp/google", VersionConstraint: "=6.15.0"}
	steps[0].ExternalProviders["google"] = googleProvider
	steps[1].ExternalProviders = map[string]resource.ExternalProvider{"google": googleProvider}
	steps[1].ConfigPlanChecks.PreApply = []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: acc.TestAccCheckGCPPrivatelinkDestroyed,
		Steps:        steps,
	})
}

func testAccGCPPrivatelinkConnectionApprovalResource() string {
	return `
resource "aiven_gcp_privatelink_connection_approval" "endpoint" {
  project           = aiven_gcp_privatelink.foo.project
  service_name      = aiven_gcp_privatelink.foo.service_name
  user_ip_address   = google_compute_address.endpoint.address
  psc_connection_id = google_compute_forwarding_rule.endpoint.psc_connection_id
}
`
}

func approvalConfigWithoutSelector(ip string) string {
	return fmt.Sprintf(`
resource "aiven_gcp_privatelink_connection_approval" "test" {
  project         = "test-project"
  service_name    = "test-service"
  user_ip_address = %q
}
`, ip)
}

func approvalConfigWithSelector(ip, psc string) string {
	return fmt.Sprintf(`
resource "aiven_gcp_privatelink_connection_approval" "test" {
  project           = "test-project"
  service_name      = "test-service"
  user_ip_address   = %q
  psc_connection_id = %q
}
`, ip, psc)
}

func frameworkFactories(client avngen.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
	}
}

// Terraform exercises replacement, immutable IP diagnostics and state-only deletion
// against a saved API snapshot. No cloud credentials or services are needed.
func TestApprovalTerraformLifecycle(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		remote := &approvalRemote{connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{{
			PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypePendingUserApproval,
		}}}
		client := remote.client(t)
		first := approvalConfigWithoutSelector("10.0.0.2")
		selected := approvalConfigWithSelector("10.0.0.2", firstPSC)
		replacement := approvalConfigWithSelector("10.0.0.3", secondPSC)
		var identityImport acc.IdentityImportVersionCheck
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: frameworkFactories(client),
			TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
			Steps: []resource.TestStep{
				{
					Config: first,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "id", "test-project/test-service"),
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", firstPSC),
						resource.TestCheckResourceAttr(approvalAddress, "state", "user-approved"),
						resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
					),
				},
				{
					ResourceName: approvalAddress, ImportState: true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
				},
				{
					PreConfig: func() {
						remote.mu.Lock()
						defer remote.mu.Unlock()
						remote.connections = append(remote.connections, privatelink.ServicePrivatelinkGoogleConnectionListOut{
							PrivatelinkConnectionId: "plc2", PscConnectionId: secondPSC, State: privatelink.ConnectionStateTypePendingUserApproval,
						})
					},
					Config:           first,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
				{
					Config:           selected,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
				{
					Config:      approvalConfigWithSelector("10.0.0.9", firstPSC),
					ExpectError: regexp.MustCompile(`endpoint IP cannot be\s+set or changed`),
				},
				{
					Config: replacement,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(approvalAddress, plancheck.ResourceActionDestroyBeforeCreate),
					}},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc2"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", secondPSC),
						resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.3"),
					),
				},
				{
					ResourceName: approvalAddress, ImportState: true, ImportStateId: "test-project/test-service/" + secondPSC,
					ImportStateVerify: true,
				},
				{
					ResourceName: approvalAddress, ImportState: true, ImportStateId: "test-project/test-service",
					ExpectError: regexp.MustCompile(`set\s+psc_connection_id`),
				},
				{
					Config:           replacement,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
			},
		})
		remote.mu.Lock()
		defer remote.mu.Unlock()
		require.Equal(t, []string{"plc1", "plc2"}, remote.approved)
		require.Len(t, remote.connections, 2, "destroying approval must preserve both consumer connections")
		require.Equal(t, "10.0.0.2", remote.connections[0].UserIPAddress)
		require.Equal(t, "10.0.0.3", remote.connections[1].UserIPAddress)
	})

	t.Run("empty selector", func(t *testing.T) {
		config := approvalConfigWithSelector("10.0.0.2", "")
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(approvalAddress, "id", "test-project/test-service"),
			resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
			resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", ""),
			resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
		)

		t.Run("new approval", func(t *testing.T) {
			remote := &approvalRemote{connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{{
				PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypePendingUserApproval,
			}}}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: frameworkFactories(remote.client(t)),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							checks,
							resource.TestCheckResourceAttr(approvalAddress, "state", "user-approved"),
						),
					},
					{
						Config:           config,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
						Check:            checks,
					},
				},
			})
		})

		t.Run("existing approval", func(t *testing.T) {
			client := avngen.NewMockClient(t)
			client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, "test-project", "test-service").Return(nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{{
					PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.2",
				}}, nil)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: frameworkFactories(client),
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: resource.ComposeTestCheckFunc(
							checks,
							resource.TestCheckResourceAttr(approvalAddress, "state", "active"),
						),
					},
					{
						Config:           config,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
						Check:            checks,
					},
				},
			})
		})
	})

	t.Run("stale read after create", func(t *testing.T) {
		config := approvalConfigWithSelector("10.0.0.2", firstPSC)
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: frameworkFactories(approvalClientWithStaleRead(t)),
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", firstPSC),
						resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
						resource.TestCheckResourceAttr(approvalAddress, "state", "user-approved"),
					),
				},
				{
					Config:           config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
			},
		})
	})

	t.Run("selector unknown until apply", func(t *testing.T) {
		remote := &approvalRemote{connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{
			{PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypePendingUserApproval},
			{PrivatelinkConnectionId: "plc2", PscConnectionId: secondPSC, State: privatelink.ConnectionStateTypePendingUserApproval},
		}}
		config := func(psc string) string {
			return fmt.Sprintf(`
resource "terraform_data" "endpoint" {
  input = %q
}
resource "aiven_gcp_privatelink_connection_approval" "test" {
  project           = "test-project"
  service_name      = "test-service"
  user_ip_address   = "10.0.0.2"
  psc_connection_id = terraform_data.endpoint.output
}
`, psc)
		}
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: frameworkFactories(remote.client(t)),
			TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_4_0)},
			Steps: []resource.TestStep{
				{
					Config: config(firstPSC),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", firstPSC),
					),
				},
				{
					Config: config(secondPSC),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(approvalAddress, plancheck.ResourceActionDestroyBeforeCreate),
					}},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc2"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", secondPSC),
					),
				},
			},
		})
	})
}

// An imported pending connection has no IP. Applying its configuration is the
// supported Update path; it must record the first approval rather than replace it.
func TestApprovalTerraformUpdatePendingImport(t *testing.T) {
	config := approvalConfigWithSelector("10.0.0.2", firstPSC)

	t.Run("basic", func(t *testing.T) {
		remote := &approvalRemote{connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{{
			PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypePendingUserApproval,
		}}}
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: frameworkFactories(remote.client(t)),
			TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_5_0)},
			Steps: []resource.TestStep{
				{
					Config: config + `
import {
  to = aiven_gcp_privatelink_connection_approval.test
  id = "test-project/test-service/50079288774361602"
}
`,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(approvalAddress, plancheck.ResourceActionUpdate),
					}},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
						resource.TestCheckResourceAttr(approvalAddress, "state", "user-approved"),
					),
				},
				{
					Config:           config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
			},
		})
		remote.mu.Lock()
		defer remote.mu.Unlock()
		require.Equal(t, []string{"plc1"}, remote.approved)
	})

	t.Run("stale read after update", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: frameworkFactories(approvalClientWithStaleRead(t)),
			TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_5_0)},
			Steps: []resource.TestStep{
				{
					Config: config + `
import {
  to = aiven_gcp_privatelink_connection_approval.test
  id = "test-project/test-service/50079288774361602"
}
`,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(approvalAddress, plancheck.ResourceActionUpdate),
					}},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", firstPSC),
						resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
						resource.TestCheckResourceAttr(approvalAddress, "state", "user-approved"),
					),
				},
				{
					Config:           config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
			},
		})
	})
}

func TestApprovalTerraformSDKState(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		var rediscovered atomic.Bool
		client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").
			RunAndReturn(func(context.Context, string, string) ([]privatelink.ServicePrivatelinkGoogleConnectionListOut, error) {
				id := "plc1"
				if rediscovered.Load() {
					id = "plc-new"
				}
				return []privatelink.ServicePrivatelinkGoogleConnectionListOut{{
					PrivatelinkConnectionId: id, PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.2",
				}}, nil
			})
		current := frameworkFactories(client)
		config := `
resource "aiven_gcp_privatelink_connection_approval" "test" {
  project         = "test-project"
  service_name    = "test-service"
  user_ip_address = "10.0.0.2"
  timeouts {
    read = "1m"
  }
}
`
		resource.UnitTest(t, resource.TestCase{
			Steps: []resource.TestStep{
				{ProtoV6ProviderFactories: sdkFactories(t), Config: config},
				{
					ProtoV6ProviderFactories: current,
					Config:                   config,
					ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(approvalAddress, "id", "test-project/test-service"),
						resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
						resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", firstPSC),
						resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
						resource.TestCheckResourceAttr(approvalAddress, "timeouts.read", "1m"),
					),
				},
				{
					ProtoV6ProviderFactories: current,
					ResourceName:             approvalAddress, ImportState: true, ImportStateVerify: true,
					ImportStateId: "test-project/test-service", ImportStateVerifyIgnore: []string{"timeouts"},
				},
				{
					ProtoV6ProviderFactories: current,
					ResourceName:             approvalAddress, ImportState: true, ImportStateVerify: true,
					ImportStateId: "test-project/test-service/" + firstPSC, ImportStateVerifyIgnore: []string{"timeouts"},
				},
				{
					ProtoV6ProviderFactories: current,
					PreConfig:                func() { rediscovered.Store(true) },
					Config:                   config,
					ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
					Check:                    resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc-new"),
				},
			},
		})
	})

	t.Run("empty selector", func(t *testing.T) {
		// SDKv2 stored the discovered PSC ID even when configuration explicitly used "".
		// A second endpoint must not make this upgrade forget the selected connection.
		client := avngen.NewMockClient(t)
		client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").Return(
			[]privatelink.ServicePrivatelinkGoogleConnectionListOut{
				{PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.2"},
				{PrivatelinkConnectionId: "plc2", PscConnectionId: secondPSC, State: privatelink.ConnectionStateTypeActive, UserIPAddress: "10.0.0.3"},
			}, nil)
		current := frameworkFactories(client)
		config := approvalConfigWithSelector("10.0.0.2", "")
		checks := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(approvalAddress, "id", "test-project/test-service"),
			resource.TestCheckResourceAttr(approvalAddress, "privatelink_connection_id", "plc1"),
			resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", ""),
			resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
			resource.TestCheckResourceAttr(approvalAddress, "state", "active"),
		)
		resource.UnitTest(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: sdkFactories(t),
					Config:                   config,
					Check:                    resource.TestCheckResourceAttr(approvalAddress, "psc_connection_id", firstPSC),
				},
				{
					ProtoV6ProviderFactories: current,
					Config:                   config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(approvalAddress, plancheck.ResourceActionUpdate),
					}},
					Check: checks,
				},
				{
					ProtoV6ProviderFactories: current,
					Config:                   config,
					ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
					Check:                    checks,
				},
			},
		})
	})
}

func TestApprovalTerraformValidatesIP(t *testing.T) {
	t.Run("plan", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: frameworkFactories(avngen.NewMockClient(t)),
			Steps: []resource.TestStep{
				{
					Config: approvalConfigWithSelector("999.1.1.1", firstPSC), PlanOnly: true,
					ExpectError: regexp.MustCompile("must be a valid IPv4 address"),
				},
				{
					Config: approvalConfigWithSelector("::1", firstPSC), PlanOnly: true,
					ExpectError: regexp.MustCompile("must be a valid IPv4 address"),
				},
				{
					Config: approvalConfigWithSelector("8.8.8.8", firstPSC), PlanOnly: true, ExpectNonEmptyPlan: true,
				},
			},
		})
	})

	t.Run("apply", func(t *testing.T) {
		config := func(ip string) string {
			return fmt.Sprintf(`
resource "terraform_data" "endpoint_ip" {
  input = %q
}
resource "aiven_gcp_privatelink_connection_approval" "test" {
  project           = "test-project"
  service_name      = "test-service"
  user_ip_address   = terraform_data.endpoint_ip.output
  psc_connection_id = %q
}
`, ip, firstPSC)
		}

		t.Run("IPv4", func(t *testing.T) {
			remote := &approvalRemote{connections: []privatelink.ServicePrivatelinkGoogleConnectionListOut{{
				PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypePendingUserApproval,
			}}}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: frameworkFactories(remote.client(t)),
				TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_4_0)},
				Steps: []resource.TestStep{{
					Config: config("10.0.0.2"),
					Check:  resource.TestCheckResourceAttr(approvalAddress, "user_ip_address", "10.0.0.2"),
				}},
			})
		})

		t.Run("IPv6", func(t *testing.T) {
			// Invalid input must fail schema validation before any API call.
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: frameworkFactories(avngen.NewMockClient(t)),
				TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_4_0)},
				Steps: []resource.TestStep{{
					Config:      config("::1"),
					ExpectError: regexp.MustCompile("must be a valid IPv4 address"),
				}},
			})
		})
	})
}

// The SDK fixture emits the original resource's state, including legacy timeouts.
// Only the current Framework provider performs subsequent reads and imports.
func sdkFactories(t *testing.T) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	legacy := &schema.Provider{ResourcesMap: map[string]*schema.Resource{
		"aiven_gcp_privatelink_connection_approval": {
			Schema: map[string]*schema.Schema{
				"project":                   {Type: schema.TypeString, Required: true, ForceNew: true},
				"service_name":              {Type: schema.TypeString, Required: true, ForceNew: true},
				"user_ip_address":           {Type: schema.TypeString, Required: true},
				"privatelink_connection_id": {Type: schema.TypeString, Computed: true},
				"psc_connection_id":         {Type: schema.TypeString, Optional: true, Computed: true, ForceNew: true},
				"state":                     {Type: schema.TypeString, Computed: true},
			},
			Timeouts: schemautil.DefaultResourceTimeouts(),
			CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) sdkdiag.Diagnostics {
				d.SetId("test-project/test-service")
				for key, value := range map[string]string{"privatelink_connection_id": "plc1", "psc_connection_id": firstPSC, "state": "active"} {
					if err := d.Set(key, value); err != nil {
						return sdkdiag.FromErr(err)
					}
				}
				return nil
			},
			ReadContext:   func(context.Context, *schema.ResourceData, any) sdkdiag.Diagnostics { return nil },
			UpdateContext: func(context.Context, *schema.ResourceData, any) sdkdiag.Diagnostics { return nil },
			DeleteContext: func(context.Context, *schema.ResourceData, any) sdkdiag.Diagnostics { return nil },
		},
	}}
	return map[string]func() (tfprotov6.ProviderServer, error){
		"aiven": func() (tfprotov6.ProviderServer, error) {
			return tf5to6server.UpgradeServer(t.Context(), legacy.GRPCProvider)
		},
	}
}

// A concurrent GET can capture the pending state before approval and be shared
// with the post-Create/Update Read by the generated client's singleflight.
func approvalClientWithStaleRead(t *testing.T) *avngen.MockClient {
	t.Helper()
	client := avngen.NewMockClient(t)
	pending := []privatelink.ServicePrivatelinkGoogleConnectionListOut{{
		PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC, State: privatelink.ConnectionStateTypePendingUserApproval,
	}}
	approved := &privatelink.ServicePrivatelinkGoogleConnectionCreateOut{
		PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC,
		State: privatelink.ServicePrivatelinkGoogleConnectionStateTypeUserApproved, UserIPAddress: "10.0.0.2",
	}
	client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, "test-project", "test-service").Return(nil).Once()
	discovery := client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").Return(pending, nil)
	client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, "test-project", "test-service", "plc1",
		&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: "10.0.0.2"}).
		Run(func(context.Context, string, string, string, *privatelink.ServicePrivatelinkGoogleConnectionCreateIn) {
			discovery.Unset()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").Return(pending, nil).Once()
			client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").
				Return([]privatelink.ServicePrivatelinkGoogleConnectionListOut{{
					PrivatelinkConnectionId: "plc1", PscConnectionId: firstPSC,
					State: privatelink.ConnectionStateTypeUserApproved, UserIPAddress: "10.0.0.2",
				}}, nil)
		}).Return(approved, nil).Once()
	return client
}

type approvalRemote struct {
	mu          sync.Mutex
	connections []privatelink.ServicePrivatelinkGoogleConnectionListOut
	approved    []string
}

func (r *approvalRemote) client(t *testing.T) *avngen.MockClient {
	t.Helper()
	client := avngen.NewMockClient(t)
	client.EXPECT().ServicePrivatelinkGoogleRefresh(mock.Anything, "test-project", "test-service").Return(nil)
	client.EXPECT().ServicePrivatelinkGoogleConnectionList(mock.Anything, "test-project", "test-service").
		RunAndReturn(func(context.Context, string, string) ([]privatelink.ServicePrivatelinkGoogleConnectionListOut, error) {
			r.mu.Lock()
			defer r.mu.Unlock()
			return append([]privatelink.ServicePrivatelinkGoogleConnectionListOut(nil), r.connections...), nil
		})
	client.EXPECT().ServicePrivatelinkGoogleConnectionCreate(mock.Anything, "test-project", "test-service", mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _, id string, req *privatelink.ServicePrivatelinkGoogleConnectionCreateIn) (*privatelink.ServicePrivatelinkGoogleConnectionCreateOut, error) {
			r.mu.Lock()
			defer r.mu.Unlock()
			for i := range r.connections {
				conn := &r.connections[i]
				if conn.PrivatelinkConnectionId != id {
					continue
				}
				if conn.State != privatelink.ConnectionStateTypePendingUserApproval {
					return nil, avngen.Error{Status: http.StatusConflict}
				}
				conn.State, conn.UserIPAddress = privatelink.ConnectionStateTypeUserApproved, req.UserIPAddress
				r.approved = append(r.approved, id)
				return &privatelink.ServicePrivatelinkGoogleConnectionCreateOut{
					PrivatelinkConnectionId: id, PscConnectionId: conn.PscConnectionId,
					State: privatelink.ServicePrivatelinkGoogleConnectionStateType(conn.State), UserIPAddress: conn.UserIPAddress,
				}, nil
			}
			return nil, avngen.Error{Status: http.StatusNotFound}
		})
	return client
}
