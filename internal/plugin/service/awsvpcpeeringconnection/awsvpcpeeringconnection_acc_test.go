package awsvpcpeeringconnection_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	peeringResource      = "aiven_aws_vpc_peering_connection.peering_connection"
	awsVPCPeeringTestEnv = "AWS_VPC_PEERING_TEST"
)

var awsProvider = resource.ExternalProvider{Source: "hashicorp/aws", VersionConstraint: "=4.40.0"}

// Exercise legacy IDs, both import protocols and provider rollback without cloud credentials.
func TestAivenAWSVPCPeeringConnection_legacyIDAndIdentity(t *testing.T) {
	t.Run("four-part ID", func(t *testing.T) { testLegacyIDAndIdentity(t, false) })
	t.Run("five-part ID", func(t *testing.T) { testLegacyIDAndIdentity(t, true) })
}

func testLegacyIDAndIdentity(t *testing.T, includeRegion bool) {
	t.Helper()
	const vpcID = "12345678-1234-4234-8234-123456789012"
	const fourPartID = "project/" + vpcID + "/123456789012/vpc-0123456789abcdef0"
	const fullID = fourPartID + "/eu-west-1"
	oldID := fourPartID
	if includeRegion {
		oldID = fullID
	}
	const config = `
resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = "project/12345678-1234-4234-8234-123456789012"
  aws_account_id = "123456789012"
  aws_vpc_id     = "vpc-0123456789abcdef0"
  aws_vpc_region = "eu-west-1"
}`
	legacy := &schema.Provider{ResourcesMap: map[string]*schema.Resource{
		"aiven_aws_vpc_peering_connection": {
			Schema: map[string]*schema.Schema{
				"vpc_id":                        {Type: schema.TypeString, Required: true, ForceNew: true},
				"aws_account_id":                {Type: schema.TypeString, Required: true, ForceNew: true},
				"aws_vpc_id":                    {Type: schema.TypeString, Required: true, ForceNew: true},
				"aws_vpc_region":                {Type: schema.TypeString, Required: true, ForceNew: true},
				"state":                         {Type: schema.TypeString, Computed: true},
				"aws_vpc_peering_connection_id": {Type: schema.TypeString, Computed: true},
			},
			CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) sdkdiag.Diagnostics {
				d.SetId(oldID)
				if err := d.Set("state", "PENDING_PEER"); err != nil {
					return sdkdiag.FromErr(err)
				}
				if err := d.Set("aws_vpc_peering_connection_id", "pcx-old"); err != nil {
					return sdkdiag.FromErr(err)
				}
				return nil
			},
			ReadContext: func(context.Context, *schema.ResourceData, any) sdkdiag.Diagnostics { return nil },
			DeleteContext: func(_ context.Context, d *schema.ResourceData, _ any) sdkdiag.Diagnostics {
				d.SetId("")
				return nil
			},
		},
	}}
	client := avngen.NewMockClient(t)
	var mu sync.Mutex
	deleted := false
	client.EXPECT().VpcGet(mock.Anything, "project", vpcID).RunAndReturn(func(context.Context, string, string) (*vpc.VpcGetOut, error) {
		mu.Lock()
		defer mu.Unlock()
		result := &vpc.VpcGetOut{}
		if !deleted {
			result.PeeringConnections = []vpc.PeeringConnectionOut{{
				PeerCloudAccount: "123456789012", PeerVpc: "vpc-0123456789abcdef0", PeerRegion: new("eu-west-1"),
				State: vpc.VpcPeeringConnectionStateTypeActive,
			}}
		}
		return result, nil
	})
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(mock.Anything, "project", vpcID, "123456789012", "vpc-0123456789abcdef0", "eu-west-1").
		RunAndReturn(func(context.Context, string, string, string, string, string) (*vpc.VpcPeeringConnectionWithRegionDeleteOut, error) {
			mu.Lock()
			defer mu.Unlock()
			deleted = true
			return &vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil
		}).Once()
	var useFramework atomic.Bool
	var identityImport acc.IdentityImportVersionCheck
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"aiven": func() (tfprotov6.ProviderServer, error) {
			if !useFramework.Load() {
				return tf5to6server.UpgradeServer(t.Context(), legacy.GRPCProvider)
			}
			return providerserver.NewProtocol6(&plugin.AivenProvider{GenClient: client})(), nil
		},
	}
	identityCheck := statecheck.ExpectIdentity(peeringResource, map[string]knownvalue.Check{
		"project": knownvalue.StringExact("project"), "project_vpc_id": knownvalue.StringExact(vpcID),
		"aws_account_id": knownvalue.StringExact("123456789012"), "aws_vpc_id": knownvalue.StringExact("vpc-0123456789abcdef0"),
		"aws_vpc_region": knownvalue.StringExact("eu-west-1"),
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
		AdditionalCLIOptions:     &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr(peeringResource, "id", oldID),
			},
			{
				PreConfig: func() { useFramework.Store(true) },
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(peeringResource, "id", oldID),
					resource.TestCheckResourceAttr(peeringResource, "project", "project"),
					resource.TestCheckResourceAttr(peeringResource, "project_vpc_id", vpcID),
					resource.TestCheckResourceAttr(peeringResource, "state", "ACTIVE"),
					resource.TestCheckNoResourceAttr(peeringResource, "aws_vpc_peering_connection_id"),
				),
			},
			{
				RefreshState: true,
			},
			{
				Config:           config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(peeringResource, "id", oldID),
					resource.TestCheckResourceAttr(peeringResource, "state", "ACTIVE"),
					resource.TestCheckNoResourceAttr(peeringResource, "aws_vpc_peering_connection_id"),
				),
			},
			{
				// Terraform 1.12/1.13 cannot read saved identity with an identity-less provider.
				// Fixed in Terraform 1.14: https://github.com/hashicorp/terraform/pull/37709.
				PreConfig: func() {
					skip, err := identityImport.SkipProviderRollback()
					require.NoError(t, err)
					useFramework.Store(skip)
				},
				SkipFunc:         identityImport.SkipProviderRollback,
				Config:           config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            resource.TestCheckResourceAttr(peeringResource, "id", oldID),
			},
			{
				PreConfig: func() { useFramework.Store(true) },
				Config:    config,
				Check:     resource.TestCheckResourceAttr(peeringResource, "id", oldID),
			},
			{
				ResourceName: peeringResource, ImportState: true, ImportStateId: fourPartID, ImportStateVerify: true,
				// Match by the peer VPC because the two string IDs identify the same connection.
				ImportStateVerifyIdentifierAttribute: "aws_vpc_id",
				ImportStateVerifyIgnore:              []string{"id"},
			},
			{
				ResourceName: peeringResource, ImportState: true, ImportStateId: fullID, ImportStateVerify: true,
				ImportStateVerifyIdentifierAttribute: "aws_vpc_id",
				ImportStateVerifyIgnore:              []string{"id"},
			},
			{
				Config:            config,
				SkipFunc:          identityImport.Skip,
				ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				ConfigStateChecks: []statecheck.StateCheck{identityCheck},
			},
			{
				// Forget the resource without deleting it so the next apply must import it.
				Config: `
removed {
  from = aiven_aws_vpc_peering_connection.peering_connection
  lifecycle {
    destroy = false
  }
}`,
				SkipFunc: identityImport.Skip,
				Check: func(state *terraform.State) error {
					if _, exists := state.RootModule().Resources[peeringResource]; exists {
						return fmt.Errorf("%s was not removed from state", peeringResource)
					}
					return nil
				},
			},
			{
				Config: config + `
import {
  to = aiven_aws_vpc_peering_connection.peering_connection
  identity = {
    project        = "project"
    project_vpc_id = "12345678-1234-4234-8234-123456789012"
    aws_account_id = "123456789012"
    aws_vpc_id     = "vpc-0123456789abcdef0"
    aws_vpc_region = "eu-west-1"
  }
}`,
				SkipFunc: identityImport.Skip,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionNoop),
					plancheck.ExpectKnownValue(peeringResource, tfjsonpath.New("id"), knownvalue.StringExact(fullID)),
				}},
				ConfigStateChecks: []statecheck.StateCheck{identityCheck},
				Check:             resource.TestCheckResourceAttr(peeringResource, "id", fullID),
			},
			{
				Config:           config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
		},
	})
}

type awsConfig struct {
	project string
	region  string
	prefix  string
	vpcID   string // Composite ID of a shared VPC; empty when Terraform owns the VPC.
}

func TestAccAivenAWSVPCPeeringConnection_basic(t *testing.T) {
	c := getAWSConfig(t)
	config := awsPeeringConfig(c, c.region)
	var peeringID string
	var identityImport acc.IdentityImportVersionCheck
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
		CheckDestroy:             checkAWSPeeringDestroy,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(awsPeeringChecks(c),
					resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", c.region),
					func(state *terraform.State) error {
						peeringID = state.RootModule().Resources[peeringResource].Primary.ID
						return nil
					},
				),
			},
			{ResourceName: peeringResource, ImportState: true, ImportStateVerify: true},
			{
				ResourceName:    peeringResource,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				SkipFunc:        identityImport.Skip,
				// The test framework checks that identity is preserved and import plans no changes.
				ImportPlanChecks: resource.ImportPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue(peeringResource, tfjsonpath.New("id"), knownvalue.StringFunc(func(value string) error {
						if value != peeringID {
							return fmt.Errorf("expected imported peering ID %q, got %q", peeringID, value)
						}
						return nil
					})),
				}},
			},
			{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{
				// Keep both VPCs so parent deletion cannot hide a broken peering Delete.
				Config: awsFixtureConfig(c),
				Check:  func(_ *terraform.State) error { return checkAWSPeeringDeleted(t.Context(), peeringID) },
			},
		},
	})
}

func TestAccAivenAWSVPCPeeringConnection_sharedVPC(t *testing.T) {
	acc.SkipIfNotAcc(t)
	c := getAWSConfig(t)
	acc.TestAccPreCheck(t)
	fixture := createAWSPeeringVPCFixture(t, c.project)
	healthy := true
	run := func(name string, test func(*testing.T, awsConfig)) {
		t.Run(name, func(t *testing.T) {
			if !healthy {
				t.Skip("shared VPC is no longer ready or contains connections left by a previous scenario")
			}
			started := time.Now()
			t.Cleanup(func() {
				t.Logf("scenario including cleanup: %s", time.Since(started).Round(time.Second))
			})
			// Registered before the scenario so its Terraform destroy and extra-row
			// cleanup finish before we inspect the still-live parent.
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				if err := fixture.checkEmpty(ctx); err != nil {
					healthy = false
					t.Errorf("shared VPC is not reusable: %v", err)
				}
			})
			c := getAWSConfig(t)
			c.vpcID = schemautil.BuildResourceID(fixture.project, fixture.id)
			test(t, c)
		})
	}

	// Each scenario has its own Terraform state and peer VPC. Run sequentially
	// so a failed cleanup cannot affect another scenario using the shared parent.
	run("backwardCompat", testAccAWSPeeringBackwardCompat)
	run("backwardCompatWithoutRefreshRejected", testAccAWSPeeringBackwardCompatWithoutRefreshRejected)
	run("createWithOtherRegion", testAccAWSPeeringCreateWithOtherRegion)
	run("replaceRegion", testAccAWSPeeringReplaceRegion)
}

func testAccAWSPeeringBackwardCompat(t *testing.T, c awsConfig) {
	c.region = "eu-west-1"
	var identityImport acc.IdentityImportVersionCheck
	steps := awsBackwardCompatibilitySteps(t, c)
	// SDK configurations could omit the region's value. Before upgrading, set
	// the region that the SDK has already recorded in state; no replacement is needed.
	steps[0].Config = awsPeeringConfig(c, "")
	steps[0].Check = resource.ComposeTestCheckFunc(steps[0].Check,
		resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", c.region),
	)
	steps[1].ConfigPlanChecks.PreApply = []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}
	steps = append(steps, acc.ProviderRollbackSteps(t, steps[0], steps[1], &identityImport)...)
	resource.Test(t, resource.TestCase{
		PreCheck:               func() { acc.TestAccPreCheck(t) },
		CheckDestroy:           checkAWSPeeringDestroy,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
		Steps:                  steps,
	})
}

func TestAccAivenAWSVPCPeeringConnection_backwardCompatWithoutRefresh(t *testing.T) {
	c := getAWSConfig(t)
	var peeringID, awsPeeringID string
	var activeStateInfo map[string]string
	steps := awsBackwardCompatibilitySteps(t, c)
	steps[0].Check = resource.ComposeTestCheckFunc(steps[0].Check,
		resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
		resource.TestCheckResourceAttrSet(peeringResource, "aws_vpc_peering_connection_id"),
		func(state *terraform.State) error {
			r := state.RootModule().Resources[peeringResource]
			awsPeeringID = r.Primary.Attributes["aws_vpc_peering_connection_id"]
			peeringID = r.Primary.ID
			return nil
		},
	)
	steps[1].PreConfig = func() {
		// Accept outside Terraform so its SDK state still says PENDING_PEER during migration.
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "aws", "ec2", "accept-vpc-peering-connection",
			"--vpc-peering-connection-id", awsPeeringID, "--region", c.region, "--no-cli-pager", "--no-cli-auto-prompt",
			"--query", "VpcPeeringConnection.Status", "--output", "json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("accept AWS peering connection: %v\n%s", err, output)
		}
		t.Logf("AWS accept response for %s in %s: %s", awsPeeringID, c.region, output)
		id, err := pluginvpc.ParseProjectPeeringID(peeringID)
		if err != nil {
			t.Fatal(err)
		}
		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			t.Fatal(err)
		}
		var lastConnection atomic.Pointer[vpc.PeeringConnectionOut]
		wait := &retry.StateChangeConf{
			Pending: []string{"APPROVED", "APPROVED_PEER_REQUESTED", "PENDING_PEER"},
			Target:  []string{"ACTIVE"},
			Refresh: func() (any, string, error) {
				connection, err := id.Find(ctx, client)
				if err != nil {
					return nil, "", err
				}
				if previous := lastConnection.Swap(connection); previous == nil || previous.State != connection.State {
					t.Logf("Aiven peering %s: state=%s, state_info=%v", id, connection.State, pluginvpc.StateInfoMap(connection.StateInfo))
				}
				return connection, string(connection.State), nil
			},
			Timeout:      10 * time.Minute,
			PollInterval: 5 * time.Second,
		}
		connection, err := wait.WaitForStateContext(ctx)
		if err != nil {
			if last := lastConnection.Load(); last != nil {
				t.Logf("Last Aiven response for %s: state=%s, state_info=%v", id, last.State, pluginvpc.StateInfoMap(last.StateInfo))
			} else {
				t.Logf("No successful Aiven peering lookup for %s", id)
			}
			// The wait context may have expired; use a separate bounded context for diagnostics.
			diagnosticCtx, diagnosticCancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer diagnosticCancel()
			cmd := exec.CommandContext(diagnosticCtx, "aws", "ec2", "describe-vpc-peering-connections",
				"--vpc-peering-connection-ids", awsPeeringID, "--region", c.region, "--no-cli-pager", "--no-cli-auto-prompt",
				"--query", "VpcPeeringConnections[].Status", "--output", "json")
			output, describeErr := cmd.CombinedOutput()
			t.Logf("AWS peering %s after waiting: %s (error: %v)", awsPeeringID, output, describeErr)
			t.Fatalf("wait for Aiven to report ACTIVE peering: %v", err)
		}
		activeStateInfo = pluginvpc.StateInfoMap(connection.(*vpc.PeeringConnectionOut).StateInfo)
	}
	steps[1].ConfigPlanChecks = resource.ConfigPlanChecks{
		PreApply:            []plancheck.PlanCheck{plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionUpdate)},
		PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
	steps[1].Check = resource.ComposeTestCheckFunc(steps[1].Check,
		resource.TestCheckResourceAttr(peeringResource, "state", "ACTIVE"),
		resource.TestCheckResourceAttr(peeringResource, "project", c.project),
		func(state *terraform.State) error {
			// The unchanged parent VPC is not refreshed, so its separate project_vpc_id may still be absent.
			_, vpcID, err := schemautil.SplitResourceID2(state.RootModule().Resources["aiven_project_vpc.aiven_vpc"].Primary.ID)
			if err != nil {
				return err
			}
			return resource.TestCheckResourceAttr(peeringResource, "project_vpc_id", vpcID)(state)
		},
		resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", c.region),
		func(state *terraform.State) error {
			return resource.TestCheckResourceAttr(peeringResource, "id", peeringID)(state)
		},
		func(state *terraform.State) error {
			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(peeringResource, "aws_vpc_peering_connection_id", awsPeeringID),
				resource.TestCheckResourceAttr(peeringResource, "state_info.%", strconv.Itoa(len(activeStateInfo))),
			}
			for key, value := range activeStateInfo {
				checks = append(checks, resource.TestCheckResourceAttr(peeringResource, "state_info."+key, value))
			}
			return resource.ComposeTestCheckFunc(checks...)(state)
		},
	)
	// Verify the following normal refresh does not introduce another change.
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}
	steps = append(steps, resource.TestStep{
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
		PreConfig:                func() { options.Plan.NoRefresh = false },
		Config:                   steps[1].Config,
		ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
		Check:                    steps[1].Check,
	})
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { preCheckAWSPeeringCLI(t) },
		CheckDestroy: checkAWSPeeringDestroy,
		// The first Framework apply must migrate SDK state without a preceding Read.
		AdditionalCLIOptions: options,
		Steps:                steps,
	})
}

func testAccAWSPeeringBackwardCompatWithoutRefreshRejected(t *testing.T, c awsConfig) {
	var id *pluginvpc.ProjectPeeringID
	var cloudID string
	steps := awsBackwardCompatibilitySteps(t, c)
	steps[0].Check = resource.ComposeTestCheckFunc(
		awsPeeringChecks(c),
		resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
		resource.TestCheckResourceAttrSet(peeringResource, "aws_vpc_peering_connection_id"),
		resource.TestCheckResourceAttrPair(peeringResource, "aws_vpc_peering_connection_id", peeringResource, "state_info.aws_vpc_peering_connection_id"),
		resource.TestCheckNoResourceAttr(peeringResource, "project"),
		resource.TestCheckNoResourceAttr(peeringResource, "project_vpc_id"),
		func(state *terraform.State) error {
			r := state.RootModule().Resources[peeringResource].Primary
			cloudID = r.Attributes["aws_vpc_peering_connection_id"]
			var err error
			id, err = pluginvpc.ParseProjectPeeringID(r.ID)
			return err
		},
	)
	steps[1].PreConfig = func() {
		// Reject outside Terraform so the SDK state retains its pending status and pcx ID.
		rejectAWSPeering(t, id, cloudID)
		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			t.Fatal(err)
		}
		connection, err := id.Find(t.Context(), client)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := connection.StateInfo["aws_vpc_peering_connection_id"]; ok {
			t.Fatal("rejected peering still has an AWS peering ID in the API")
		}
	}
	steps[1].ConfigPlanChecks = resource.ConfigPlanChecks{
		PreApply:            []plancheck.PlanCheck{plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionUpdate)},
		PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
	refreshedCheck := func(state *terraform.State) error {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(peeringResource, "id", id.String()),
			resource.TestCheckResourceAttr(peeringResource, "project", id.Project),
			resource.TestCheckResourceAttr(peeringResource, "project_vpc_id", id.ProjectVPCID),
			resource.TestCheckResourceAttr(peeringResource, "state", "REJECTED_BY_PEER"),
			resource.TestCheckNoResourceAttr(peeringResource, "aws_vpc_peering_connection_id"),
			resource.TestCheckNoResourceAttr(peeringResource, "state_info.aws_vpc_peering_connection_id"),
		)(state)
	}
	steps[1].Check = refreshedCheck

	steps = append(steps,
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
			Config:                   awsPeeringCreateBeforeDestroyConfig(c, c.region),
			Taint:                    []string{peeringResource},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
				plancheck.ExpectResourceAction("aws_vpc.aws_vpc", plancheck.ResourceActionNoop),
			}},
			ExpectError: regexp.MustCompile(`already exists in state REJECTED_BY_PEER`),
		},
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
			// Check runs separately because an expected apply error skips it.
			RefreshState:       true,
			ExpectNonEmptyPlan: true, // The original resource is still tainted.
			Check: resource.ComposeTestCheckFunc(refreshedCheck,
				func(_ *terraform.State) error {
					return checkAWSPeeringState(t.Context(), id, vpc.VpcPeeringConnectionStateTypeRejectedByPeer)
				},
			),
		},
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheckAWSPeeringCLI(t) },
		CheckDestroy: checkAWSPeeringDestroy,
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: steps,
	})
}

func testAccAWSPeeringCreateWithOtherRegion(t *testing.T, c awsConfig) {
	var id, other *pluginvpc.ProjectPeeringID
	var cloudID string
	cleanup := func() {}
	checkConnection := resource.ComposeTestCheckFunc(awsPeeringChecks(c),
		resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", c.region),
		func(state *terraform.State) error {
			if err := resource.TestCheckResourceAttr(peeringResource, "id", id.String())(state); err != nil {
				return err
			}
			if err := checkAWSPeeringState(t.Context(), other, vpc.VpcPeeringConnectionStateTypeInvalidSpecification); err != nil {
				return err
			}
			return checkAWSPeeringState(t.Context(), id, vpc.VpcPeeringConnectionStateTypePendingPeer)
		},
	)
	checkCloudID := func(state *terraform.State) error {
		return resource.TestCheckResourceAttr(peeringResource, "aws_vpc_peering_connection_id", cloudID)(state)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
		CheckDestroy:             checkAWSPeeringDestroy,
		ErrorCheck: func(err error) error {
			cleanup()
			return err
		},
		Steps: []resource.TestStep{
			{
				Config: awsFixtureConfig(c),
				Check: func(state *terraform.State) error {
					var err error
					id, err = awsPeeringFixtureID(state, c.vpcID, c.region)
					return err
				},
			},
			{
				PreConfig: func() { other, cleanup = addAWSPeeringInOtherRegion(t, id) },
				Config:    awsPeeringConfig(c, c.region),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreate),
				}},
				Check: resource.ComposeTestCheckFunc(checkConnection,
					resource.TestCheckResourceAttrSet(peeringResource, "aws_vpc_peering_connection_id"),
					func(state *terraform.State) error {
						cloudID = state.RootModule().Resources[peeringResource].Primary.Attributes["aws_vpc_peering_connection_id"]
						return nil
					},
				),
			},
			{
				PreConfig:        func() { cleanup() },
				Config:           awsPeeringConfig(c, c.region),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            resource.ComposeTestCheckFunc(awsPeeringChecks(c), checkCloudID),
			},
			{
				// Add the other regional row after the resource exists.
				// Its failed cloud setup must not prevent reading the original peering.
				PreConfig:        func() { other, cleanup = addAWSPeeringInOtherRegion(t, id) },
				Config:           awsPeeringConfig(c, c.region),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            resource.ComposeTestCheckFunc(checkConnection, checkCloudID),
			},
			{
				Config: awsFixtureConfig(c),
				Check: func(_ *terraform.State) error {
					defer cleanup()
					if err := checkAWSPeeringDeleted(t.Context(), id.String()); err != nil {
						return err
					}
					return checkAWSPeeringState(t.Context(), other, vpc.VpcPeeringConnectionStateTypeInvalidSpecification)
				},
			},
		},
	})
}

// Validate both schemas with real Terraform, without cloud credentials.
func TestAivenAWSVPCPeeringConnection_requiresRegion(t *testing.T) {
	for _, block := range []string{"resource", "data"} {
		t.Run(block, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
				},
				Steps: []resource.TestStep{{
					Config: fmt.Sprintf(`
%s "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = "project/12345678-1234-4234-8234-123456789012"
  aws_account_id = "123456789012"
  aws_vpc_id     = "vpc-0123456789abcdef0"
  aws_vpc_region = ""
}`, block),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Attribute aws_vpc_region string length must be between 1 and 1024`),
				}},
			})
		})
	}
}

// Use real Terraform to verify replacement decisions and cleanup without cloud credentials.
func TestAivenAWSVPCPeeringConnection_replacement(t *testing.T) {
	const config = `
resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = "project/12345678-1234-4234-8234-123456789012"
  aws_account_id = "123456789012"
  aws_vpc_id     = "vpc-0123456789abcdef0"
  aws_vpc_region = %q
  lifecycle {
    create_before_destroy = %t
  }
}`
	const vpcID = "12345678-1234-4234-8234-123456789012"
	var mu sync.Mutex
	var connection *vpc.PeeringConnectionOut
	var operations []string
	creates := 0
	client := avngen.NewMockClient(t)
	client.EXPECT().VpcGet(mock.Anything, "project", vpcID).RunAndReturn(
		func(context.Context, string, string) (*vpc.VpcGetOut, error) {
			mu.Lock()
			defer mu.Unlock()
			result := &vpc.VpcGetOut{CloudName: "aws-eu-west-1"}
			if connection != nil {
				result.PeeringConnections = []vpc.PeeringConnectionOut{*connection}
			}
			return result, nil
		})
	client.EXPECT().VpcPeeringConnectionCreate(mock.Anything, "project", vpcID, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, request *vpc.VpcPeeringConnectionCreateIn) (*vpc.VpcPeeringConnectionCreateOut, error) {
			mu.Lock()
			defer mu.Unlock()
			operations = append(operations, "create "+*request.PeerRegion)
			creates++
			connection = &vpc.PeeringConnectionOut{
				PeerCloudAccount: request.PeerCloudAccount, PeerVpc: request.PeerVpc, PeerRegion: request.PeerRegion,
				State: vpc.VpcPeeringConnectionStateTypePendingPeer,
			}
			if creates == 1 {
				connection.State = vpc.VpcPeeringConnectionStateTypeRejectedByPeer
			}
			return &vpc.VpcPeeringConnectionCreateOut{}, nil
		})
	client.EXPECT().VpcPeeringConnectionWithRegionDelete(mock.Anything, "project", vpcID, "123456789012", "vpc-0123456789abcdef0", mock.Anything).
		RunAndReturn(func(_ context.Context, _, _, _, _, region string) (*vpc.VpcPeeringConnectionWithRegionDeleteOut, error) {
			mu.Lock()
			defer mu.Unlock()
			operations = append(operations, "delete "+region)
			if connection == nil || *connection.PeerRegion != region {
				return nil, fmt.Errorf("unexpected deletion in region %q", region)
			}
			connection = nil
			return &vpc.VpcPeeringConnectionWithRegionDeleteOut{}, nil
		})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
		},
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(config, "us-east-1", true),
				ExpectError: regexp.MustCompile(`(?s)resource reached a failed state:.*rejected`),
			},
			{
				Config: fmt.Sprintf(config, "us-east-1", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
				}},
				ExpectError: regexp.MustCompile(`already exists in state\s+REJECTED_BY_PEER`),
			},
			{
				Config: fmt.Sprintf(config, "us-east-1", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionDestroyBeforeCreate),
				}},
				Check: resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
			},
			{
				Config: fmt.Sprintf(config, "eu-west-1", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionDestroyBeforeCreate),
				}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", "eu-west-1"),
					resource.TestCheckResourceAttr(peeringResource, "id", "project/"+vpcID+"/123456789012/vpc-0123456789abcdef0/eu-west-1"),
				),
			},
			{
				// Existing resources obey the same nonempty-region requirement.
				Config:      fmt.Sprintf(config, "", false),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Attribute aws_vpc_region string length must be between 1 and 1024`),
			},
			{
				Config:           fmt.Sprintf(config, "eu-west-1", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	require.Nil(t, connection)
	require.Equal(t, []string{
		"create us-east-1", "delete us-east-1", "create us-east-1", "delete us-east-1", "create eu-west-1", "delete eu-west-1",
	}, operations)
}

func testAccAWSPeeringReplaceRegion(t *testing.T, c awsConfig) {
	// The peer VPC is in us-east-1. Correcting an explicitly wrong region
	// must replace only that failed connection.
	c.region = "us-east-1"
	var original, replacement *pluginvpc.ProjectPeeringID
	cleanupOriginal, cleanupReplacement := func() {}, func() {}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
		CheckDestroy:             checkAWSPeeringDestroy,
		// Remove both exact API keys before deferred destroy if replacement fails,
		// without hiding the original error.
		ErrorCheck: func(err error) error {
			cleanupOriginal()
			cleanupReplacement()
			return err
		},
		Steps: []resource.TestStep{
			{
				Config: awsFixtureConfig(c),
				Check: func(state *terraform.State) error {
					var err error
					original, err = awsPeeringFixtureID(state, c.vpcID, "eu-west-1")
					if err != nil {
						return err
					}
					replacement, err = awsPeeringFixtureID(state, c.vpcID, c.region)
					if err != nil {
						return err
					}
					cleanupOriginal = awsPeeringCleanup(t, original)
					cleanupReplacement = awsPeeringCleanup(t, replacement)
					return nil
				},
			},
			{
				Config:      awsPeeringCreateBeforeDestroyConfig(c, "eu-west-1"),
				ExpectError: regexp.MustCompile(`VPC\s+peering\s+connection\s+specification\s+is\s+invalid`),
			},
			{
				PreConfig: func() {
					if err := checkAWSPeeringState(t.Context(), original, vpc.VpcPeeringConnectionStateTypeInvalidSpecification); err != nil {
						t.Fatal(err)
					}
				},
				Config: awsPeeringCreateBeforeDestroyConfig(c, c.region),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
					plancheck.ExpectResourceAction("aws_vpc.aws_vpc", plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", c.region),
					resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
					resource.TestCheckResourceAttrSet(peeringResource, "aws_vpc_peering_connection_id"),
					func(_ *terraform.State) error {
						if err := checkAWSPeeringDeleted(t.Context(), original.String()); err != nil {
							return err
						}
						return checkAWSPeeringState(t.Context(), replacement, vpc.VpcPeeringConnectionStateTypePendingPeer)
					},
				),
			},
			{
				// Changing the region must replace the connection.
				// Only plan: this peer VPC does not exist in eu-west-1.
				Config:             awsPeeringCreateBeforeDestroyConfig(c, "eu-west-1"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
					plancheck.ExpectResourceAction("aws_vpc.aws_vpc", plancheck.ResourceActionNoop),
				}},
			},
			{
				Config:           awsPeeringCreateBeforeDestroyConfig(c, c.region),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
		},
	})
}

func getAWSConfig(t *testing.T) awsConfig {
	t.Helper()
	if os.Getenv(awsVPCPeeringTestEnv) == "" {
		t.Skipf("environment variable %s must be set to run this test", awsVPCPeeringTestEnv)
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-east-1"
	}
	return awsConfig{project: acc.ProjectName(), region: region, prefix: "test-tf-acc-" + acctest.RandString(7)}
}

func awsBackwardCompatibilitySteps(t *testing.T, c awsConfig) []resource.TestStep {
	t.Helper()
	steps := acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
		TFConfig:           awsPeeringConfig(c, c.region),
		OldProviderVersion: "4.62.0",
		Checks:             awsPeeringChecks(c),
	})
	for i := range steps {
		if steps[i].ExternalProviders == nil {
			steps[i].ExternalProviders = make(map[string]resource.ExternalProvider)
		}
		steps[i].ExternalProviders["aws"] = awsProvider
	}
	return steps
}

func awsPeeringChecks(c awsConfig) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("aws_vpc.aws_vpc", "cidr_block", "10.0.0.0/24"),
		resource.TestCheckResourceAttrPair(peeringResource, "aws_account_id", "aws_vpc.aws_vpc", "owner_id"),
		resource.TestCheckResourceAttrPair(peeringResource, "aws_vpc_id", "aws_vpc.aws_vpc", "id"),
		resource.TestCheckResourceAttrPair("data."+peeringResource, "id", peeringResource, "id"),
		resource.TestMatchResourceAttr(peeringResource, "state", regexp.MustCompile(`^(ACTIVE|PENDING_PEER)$`)),
		func(state *terraform.State) error {
			r := state.RootModule().Resources[peeringResource]
			accountID := state.RootModule().Resources["aws_vpc.aws_vpc"].Primary.Attributes["owner_id"]
			want := fmt.Sprintf("%s/%s/%s/%s", r.Primary.Attributes["vpc_id"], accountID, r.Primary.Attributes["aws_vpc_id"], c.region)
			if r.Primary.ID != want {
				return fmt.Errorf("expected peering ID %q, got %q", want, r.Primary.ID)
			}
			return nil
		},
	}
	if c.vpcID == "" {
		checks = append(checks,
			resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
			resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "network_cidr", "10.0.1.0/24"),
			resource.TestCheckResourceAttrPair(peeringResource, "vpc_id", "aiven_project_vpc.aiven_vpc", "id"),
		)
	} else {
		checks = append(checks, resource.TestCheckResourceAttr(peeringResource, "vpc_id", c.vpcID))
	}
	return resource.ComposeTestCheckFunc(checks...)
}

func checkAWSPeeringDeleted(ctx context.Context, value string) error {
	id, err := pluginvpc.ParseProjectPeeringID(value)
	if err != nil {
		return err
	}
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}
	connection, err := id.Find(ctx, client)
	if adapter.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if connection.State != vpc.VpcPeeringConnectionStateTypeDeleted {
		return fmt.Errorf("AWS peering %q still exists in state %s", value, connection.State)
	}
	return nil
}

func checkAWSPeeringDestroy(state *terraform.State) error {
	for _, r := range state.RootModule().Resources {
		if r.Type == "aiven_aws_vpc_peering_connection" {
			if err := checkAWSPeeringDeleted(context.Background(), r.Primary.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func preCheckAWSPeeringCLI(t *testing.T) {
	t.Helper()
	acc.TestAccPreCheck(t)
	if _, err := exec.LookPath("aws"); err != nil {
		t.Fatalf("AWS CLI v2 is required to accept or reject peering outside Terraform: %v", err)
	}
}

func awsPeeringFixtureID(state *terraform.State, compositeVPCID, region string) (*pluginvpc.ProjectPeeringID, error) {
	project, vpcID, err := schemautil.SplitResourceID2(compositeVPCID)
	if err != nil {
		return nil, err
	}
	peer := state.RootModule().Resources["aws_vpc.aws_vpc"].Primary
	return &pluginvpc.ProjectPeeringID{
		Project: project, ProjectVPCID: vpcID,
		PeerCloudAccount: peer.Attributes["owner_id"], PeerVPC: peer.ID, PeerRegion: &region,
	}, nil
}

func rejectAWSPeering(t *testing.T, id *pluginvpc.ProjectPeeringID, cloudID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "aws", "ec2", "reject-vpc-peering-connection",
		"--vpc-peering-connection-id", cloudID, "--region", *id.PeerRegion, "--no-cli-pager", "--no-cli-auto-prompt")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("reject AWS peering connection: %v\n%s", err, output)
	}
	if err := waitAWSPeeringState(ctx, id, vpc.VpcPeeringConnectionStateTypeRejectedByPeer); err != nil {
		t.Fatal(err)
	}
}

func checkAWSPeeringState(ctx context.Context, id *pluginvpc.ProjectPeeringID, want vpc.VpcPeeringConnectionStateType) error {
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}
	connection, err := id.Find(ctx, client)
	if err != nil {
		return err
	}
	if connection.State != want {
		return fmt.Errorf("AWS peering %q: expected %s in the API, got %s", id, want, connection.State)
	}
	return nil
}

func waitAWSPeeringState(ctx context.Context, id *pluginvpc.ProjectPeeringID, want vpc.VpcPeeringConnectionStateType) error {
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}
	wait := &retry.StateChangeConf{
		Target: []string{string(want)},
		Refresh: func() (any, string, error) {
			connection, err := id.Find(ctx, client)
			if adapter.IsNotFound(err) && want == vpc.VpcPeeringConnectionStateTypeDeleted {
				return id, string(want), nil
			}
			if err != nil {
				return nil, "", err
			}
			return connection, string(connection.State), nil
		},
		Timeout:      10 * time.Minute,
		PollInterval: 5 * time.Second,
	}
	if _, err := wait.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for AWS peering %q to reach %s: %w", id, want, err)
	}
	return nil
}

func addAWSPeeringInOtherRegion(t *testing.T, id *pluginvpc.ProjectPeeringID) (*pluginvpc.ProjectPeeringID, func()) {
	t.Helper()
	other := *id
	region := "us-east-1"
	if *id.PeerRegion == region {
		region = "eu-west-1"
	}
	other.PeerRegion = &region
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		t.Fatal(err)
	}
	cleanup := awsPeeringCleanup(t, &other)

	// The API accepts the region before verifying that the peer VPC exists there.
	// Terraform Create cannot be used to prepare this deliberately failed row.
	_, err = client.VpcPeeringConnectionCreate(t.Context(), other.Project, other.ProjectVPCID, &vpc.VpcPeeringConnectionCreateIn{
		PeerCloudAccount: other.PeerCloudAccount, PeerVpc: other.PeerVPC, PeerRegion: other.PeerRegion,
	})
	if err != nil {
		cleanup()
		t.Fatalf("create extra AWS peering: %v", err)
	}
	if err := waitAWSPeeringState(t.Context(), &other, vpc.VpcPeeringConnectionStateTypeInvalidSpecification); err != nil {
		cleanup()
		t.Fatal(err)
	}
	return &other, cleanup
}

func awsPeeringCleanup(t *testing.T, id *pluginvpc.ProjectPeeringID) func() {
	t.Helper()
	var cleaned bool
	cleanup := func() {
		t.Helper()
		if cleaned {
			return
		}
		// Cleanup must also work after a failed/cancelled test. The full key limits
		// deletion to this extra row in the VPC created by the test.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			t.Errorf("get client for AWS peering cleanup: %v", err)
			return
		}
		connection, err := id.Find(ctx, client)
		if adapter.IsNotFound(err) {
			cleaned = true
			return
		}
		if err != nil {
			t.Errorf("find AWS peering for cleanup: %v", err)
			return
		}
		if connection.State != vpc.VpcPeeringConnectionStateTypeDeleting {
			if err := id.Delete(ctx, client, nil); err != nil && !adapter.IsNotFound(err) {
				t.Errorf("delete AWS peering: %v", err)
				return
			}
		}
		if err := waitAWSPeeringState(ctx, id, vpc.VpcPeeringConnectionStateTypeDeleted); err != nil {
			t.Errorf("clean up AWS peering: %v", err)
			return
		}
		cleaned = true
	}
	t.Cleanup(cleanup)
	return cleanup
}

func awsFixtureConfig(c awsConfig) string {
	projectVPC := ""
	if c.vpcID == "" {
		projectVPC = fmt.Sprintf(`
resource "aiven_project_vpc" "aiven_vpc" {
  project      = %[1]q
  cloud_name   = "aws-eu-west-1"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}
`, c.project)
	}
	return fmt.Sprintf(`
provider "aws" {
  region = %[1]q
}

%[2]s

resource "aws_vpc" "aws_vpc" {
  cidr_block = "10.0.0.0/24"
  tags = {
    Name = "%[3]s-vpc-peering"
  }
}
`, c.region, projectVPC, c.prefix)
}

func (c awsConfig) vpcReference() string {
	if c.vpcID != "" {
		return strconv.Quote(c.vpcID)
	}
	return "aiven_project_vpc.aiven_vpc.id"
}

func awsPeeringConfig(c awsConfig, region string) string {
	return awsFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = %[3]s
  aws_account_id = aws_vpc.aws_vpc.owner_id
  aws_vpc_id     = aws_vpc.aws_vpc.id
  aws_vpc_region = %[1]q
}

data "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = aiven_aws_vpc_peering_connection.peering_connection.vpc_id
  aws_account_id = aiven_aws_vpc_peering_connection.peering_connection.aws_account_id
  aws_vpc_id     = aiven_aws_vpc_peering_connection.peering_connection.aws_vpc_id
  aws_vpc_region = %[2]q
}
`, region, c.region, c.vpcReference())
}

func awsPeeringCreateBeforeDestroyConfig(c awsConfig, region string) string {
	return awsFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = %[2]s
  aws_account_id = aws_vpc.aws_vpc.owner_id
  aws_vpc_id     = aws_vpc.aws_vpc.id
  aws_vpc_region = %[1]q

  lifecycle {
    create_before_destroy = true
  }
}
`, region, c.vpcReference())
}

type awsPeeringVPCFixture struct {
	client  avngen.Client
	project string
	id      string
}

func createAWSPeeringVPCFixture(t *testing.T, project string) awsPeeringVPCFixture {
	t.Helper()
	started := time.Now()
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Minute)
	defer cancel()
	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)
	response, err := client.VpcCreate(ctx, project, &vpc.VpcCreateIn{
		CloudName: "aws-eu-west-1", NetworkCidr: "10.0.1.0/24", PeeringConnections: []vpc.PeeringConnectionIn{},
	})
	require.NoError(t, err)
	fixture := awsPeeringVPCFixture{client: client, project: project, id: response.ProjectVpcId}
	t.Logf("shared VPC: %s", schemautil.BuildResourceID(project, fixture.id))
	var manifest string
	// Register deletion before waiting for readiness or writing the recovery file.
	t.Cleanup(func() {
		started := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		if err := fixture.delete(ctx); err != nil {
			t.Errorf("delete shared VPC %s/%s (recovery file %s): %v", project, fixture.id, manifest, err)
			return
		}
		if manifest != "" {
			require.NoError(t, os.Remove(manifest))
		}
		t.Logf("shared VPC cleanup: %s", time.Since(started).Round(time.Second))
	})

	// Unlike Terraform-owned resources this VPC is not in a test state. Keep its
	// exact key on disk so an interrupted test run can still be cleaned up.
	dir := os.Getenv("TF_ACC_TEMP_DIR")
	if dir != "" {
		require.NoError(t, os.MkdirAll(dir, 0o700))
	}
	file, err := os.CreateTemp(dir, "aiven-shared-vpc-*.txt")
	require.NoError(t, err)
	manifest = file.Name()
	_, err = fmt.Fprintf(file, "project=%s\nproject_vpc_id=%s\ncloud_name=aws-eu-west-1\nnetwork_cidr=10.0.1.0/24\n", project, fixture.id)
	closeErr := file.Close()
	require.NoError(t, err)
	require.NoError(t, closeErr)
	t.Logf("shared VPC recovery file: %s", manifest)

	require.NoError(t, fixture.wait(ctx, string(vpc.VpcStateTypeActive)))
	require.NoError(t, fixture.checkEmpty(ctx))
	t.Logf("shared VPC preparation: %s", time.Since(started).Round(time.Second))
	return fixture
}

func (f awsPeeringVPCFixture) wait(ctx context.Context, target string) error {
	wait := &retry.StateChangeConf{
		Target: []string{target},
		Refresh: func() (any, string, error) {
			response, err := f.client.VpcGet(ctx, f.project, f.id)
			if avngen.IsNotFound(err) {
				return f.id, "NOT_FOUND", nil
			}
			if err != nil {
				return nil, "", err
			}
			return response, string(response.State), nil
		},
		Timeout:      15 * time.Minute,
		PollInterval: 5 * time.Second,
	}
	_, err := wait.WaitForStateContext(ctx)
	return err
}

func (f awsPeeringVPCFixture) delete(ctx context.Context) error {
	_, err := f.client.VpcDelete(ctx, f.project, f.id)
	if avngen.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return f.wait(ctx, "NOT_FOUND")
}

func (f awsPeeringVPCFixture) checkEmpty(ctx context.Context) error {
	response, err := f.client.VpcGet(ctx, f.project, f.id)
	if err != nil {
		return err
	}
	if response.State != vpc.VpcStateTypeActive {
		return fmt.Errorf("expected ACTIVE VPC, got %s", response.State)
	}
	for _, connection := range response.PeeringConnections {
		if connection.State != vpc.VpcPeeringConnectionStateTypeDeleted {
			return fmt.Errorf("peering with %s/%s still exists in state %s", connection.PeerCloudAccount, connection.PeerVpc, connection.State)
		}
	}
	return nil
}

func TestAWSPeeringVPCFixture(t *testing.T) {
	t.Run("cleanup waits for the VPC to disappear", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			fixture := awsPeeringVPCFixture{client: client, project: "project", id: "vpc-id"}
			client.EXPECT().VpcDelete(mock.Anything, fixture.project, fixture.id).
				Return(&vpc.VpcDeleteOut{State: vpc.VpcStateTypeDeleting}, nil).Once()
			client.EXPECT().VpcGet(mock.Anything, fixture.project, fixture.id).
				Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeDeleting}, nil).Once()
			client.EXPECT().VpcGet(mock.Anything, fixture.project, fixture.id).
				Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeDeleted}, nil).Once()
			client.EXPECT().VpcGet(mock.Anything, fixture.project, fixture.id).
				Return(nil, avngen.Error{Status: http.StatusNotFound}).Once()

			require.NoError(t, fixture.delete(t.Context()))
		})
	})

	t.Run("delete errors are reported", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		fixture := awsPeeringVPCFixture{client: client, project: "project", id: "vpc-id"}
		forbidden := avngen.Error{Status: http.StatusForbidden}
		client.EXPECT().VpcDelete(mock.Anything, fixture.project, fixture.id).Return(nil, forbidden).Once()

		require.ErrorIs(t, fixture.delete(t.Context()), forbidden)
	})

	t.Run("remaining connections prevent reuse", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		fixture := awsPeeringVPCFixture{client: client, project: "project", id: "vpc-id"}
		client.EXPECT().VpcGet(mock.Anything, fixture.project, fixture.id).
			Return(&vpc.VpcGetOut{
				State: vpc.VpcStateTypeActive,
				PeeringConnections: []vpc.PeeringConnectionOut{{
					PeerCloudAccount: "account", PeerVpc: "peer-vpc", State: vpc.VpcPeeringConnectionStateTypeDeleting,
				}},
			}, nil).Once()
		client.EXPECT().VpcGet(mock.Anything, fixture.project, fixture.id).
			Return(&vpc.VpcGetOut{State: vpc.VpcStateTypeActive}, nil).Once()

		require.ErrorContains(t, fixture.checkEmpty(t.Context()), "peering with account/peer-vpc still exists in state DELETING")
		require.NoError(t, fixture.checkEmpty(t.Context()))
	})

	t.Run("missing parent does not count as successful cleanup", func(t *testing.T) {
		client := avngen.NewMockClient(t)
		fixture := awsPeeringVPCFixture{client: client, project: "project", id: "vpc-id"}
		notFound := avngen.Error{Status: http.StatusNotFound}
		client.EXPECT().VpcGet(mock.Anything, fixture.project, fixture.id).Return(nil, notFound).Once()

		require.ErrorIs(t, fixture.checkEmpty(t.Context()), notFound)
	})
}
