package awsvpcpeeringconnection_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

type awsConfig struct {
	project string
	region  string
	prefix  string
}

func TestAccAivenAWSVPCPeeringConnection_basic(t *testing.T) {
	c := getAWSConfig(t)
	config := awsPeeringConfig(c, c.region)
	var peeringID string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
		CheckDestroy:             checkAWSPeeringDestroy,
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
			{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{
				// Keep both VPCs so parent deletion cannot hide a broken peering Delete.
				Config: awsFixtureConfig(c),
				Check:  func(_ *terraform.State) error { return checkAWSPeeringDeleted(t.Context(), peeringID) },
			},
		},
	})
}

func TestAccAivenAWSVPCPeeringConnection_backwardCompat(t *testing.T) {
	c := getAWSConfig(t)
	c.region = "eu-west-1"
	steps := awsBackwardCompatibilitySteps(t, c)
	// SDK configurations could omit the region's value. Before upgrading, set
	// the region that the SDK has already recorded in state; no replacement is needed.
	steps[0].Config = awsPeeringConfig(c, "")
	steps[0].Check = resource.ComposeTestCheckFunc(steps[0].Check,
		resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", c.region),
	)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: checkAWSPeeringDestroy,
		Steps:        steps,
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
			"--vpc-peering-connection-id", awsPeeringID, "--region", c.region, "--no-cli-pager", "--no-cli-auto-prompt")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("accept AWS peering connection: %v\n%s", err, output)
		}
		id, err := pluginvpc.ParseProjectPeeringID(peeringID)
		if err != nil {
			t.Fatal(err)
		}
		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			t.Fatal(err)
		}
		wait := &retry.StateChangeConf{
			Pending: []string{"APPROVED", "APPROVED_PEER_REQUESTED", "PENDING_PEER"},
			Target:  []string{"ACTIVE"},
			Refresh: func() (any, string, error) {
				connection, err := id.Find(ctx, client)
				if err != nil {
					return nil, "", err
				}
				return connection, string(connection.State), nil
			},
			Timeout:      10 * time.Minute,
			PollInterval: 5 * time.Second,
		}
		connection, err := wait.WaitForStateContext(ctx)
		if err != nil {
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { preCheckAWSPeeringCLI(t) },
		CheckDestroy: checkAWSPeeringDestroy,
		// The first Framework apply must migrate SDK state without a preceding Read.
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: steps,
	})
}

func TestAccAivenAWSVPCPeeringConnection_backwardCompatWithoutRefreshRejected(t *testing.T) {
	c := getAWSConfig(t)
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
	steps[1].Check = func(state *terraform.State) error {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(peeringResource, "id", id.String()),
			resource.TestCheckResourceAttr(peeringResource, "project", id.Project),
			resource.TestCheckResourceAttr(peeringResource, "project_vpc_id", id.ProjectVPCID),
			resource.TestCheckResourceAttr(peeringResource, "state", "REJECTED_BY_PEER"),
			resource.TestCheckNoResourceAttr(peeringResource, "aws_vpc_peering_connection_id"),
			resource.TestCheckNoResourceAttr(peeringResource, "state_info.aws_vpc_peering_connection_id"),
		)(state)
	}
	steps = append(steps,
		resource.TestStep{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			ExternalProviders:        map[string]resource.ExternalProvider{"aws": awsProvider},
			Config:                   awsPeeringCreateBeforeDestroyConfig(c, c.region),
			Taint:                    []string{peeringResource},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
				plancheck.ExpectResourceAction("aiven_project_vpc.aiven_vpc", plancheck.ResourceActionNoop),
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
			Check: resource.ComposeTestCheckFunc(steps[1].Check,
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

func TestAccAivenAWSVPCPeeringConnection_createWithOtherRegion(t *testing.T) {
	c := getAWSConfig(t)
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
					id, err = awsPeeringFixtureID(state, c.region)
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

func TestAccAivenAWSVPCPeeringConnection_replaceRegion(t *testing.T) {
	c := getAWSConfig(t)
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
					original, err = awsPeeringFixtureID(state, "eu-west-1")
					if err != nil {
						return err
					}
					replacement, err = awsPeeringFixtureID(state, c.region)
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
					plancheck.ExpectResourceAction("aiven_project_vpc.aiven_vpc", plancheck.ResourceActionNoop),
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
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
					plancheck.ExpectResourceAction("aiven_project_vpc.aiven_vpc", plancheck.ResourceActionNoop),
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
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "state", "ACTIVE"),
		resource.TestCheckResourceAttr("aiven_project_vpc.aiven_vpc", "network_cidr", "10.0.1.0/24"),
		resource.TestCheckResourceAttr("aws_vpc.aws_vpc", "cidr_block", "10.0.0.0/24"),
		resource.TestCheckResourceAttrPair(peeringResource, "aws_account_id", "aws_vpc.aws_vpc", "owner_id"),
		resource.TestCheckResourceAttrPair(peeringResource, "vpc_id", "aiven_project_vpc.aiven_vpc", "id"),
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
	)
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

func awsPeeringFixtureID(state *terraform.State, region string) (*pluginvpc.ProjectPeeringID, error) {
	project, vpcID, err := schemautil.SplitResourceID2(state.RootModule().Resources["aiven_project_vpc.aiven_vpc"].Primary.ID)
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
	return fmt.Sprintf(`
provider "aws" {
  region = %[1]q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = %[2]q
  cloud_name   = "aws-eu-west-1"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aws_vpc" "aws_vpc" {
  cidr_block = "10.0.0.0/24"
  tags = {
    Name = "%[3]s-vpc-peering"
  }
}
`, c.region, c.project, c.prefix)
}

func awsPeeringConfig(c awsConfig, region string) string {
	return awsFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = aiven_project_vpc.aiven_vpc.id
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
`, region, c.region)
}

func awsPeeringCreateBeforeDestroyConfig(c awsConfig, region string) string {
	return awsFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_aws_vpc_peering_connection" "peering_connection" {
  vpc_id         = aiven_project_vpc.aiven_vpc.id
  aws_account_id = aws_vpc.aws_vpc.owner_id
  aws_vpc_id     = aws_vpc.aws_vpc.id
  aws_vpc_region = %[1]q

  lifecycle {
    create_before_destroy = true
  }
}
`, region)
}
