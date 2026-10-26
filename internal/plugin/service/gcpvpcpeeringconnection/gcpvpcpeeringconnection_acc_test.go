package gcpvpcpeeringconnection_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/avast/retry-go/v4"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
)

const (
	peeringResource      = "aiven_gcp_vpc_peering_connection.foo"
	gcpVPCPeeringTestEnv = "GCP_VPC_PEERING_TEST"
	defaultGCPRegion     = "europe-west10"
)

type gcpConfig struct {
	project     string
	peerProject string
	region      string
}

func getGCPConfig(t *testing.T) gcpConfig {
	t.Helper()
	env := acc.RequireEnvVars(t, gcpVPCPeeringTestEnv, "GOOGLE_PROJECT")
	return gcpConfig{
		project:     acc.ProjectName(),
		peerProject: env["GOOGLE_PROJECT"],
		region:      defaultGCPRegion,
	}
}

func TestAccAivenGCPPeeringConnection_basic(t *testing.T) {
	c := getGCPConfig(t)
	peeringConfig := gcpPeeringConfig(c)
	fullConfig := peeringConfig + fmt.Sprintf(`
provider "google" {
  project = %[1]q
  region  = %[2]q
}

data "google_compute_network" "foo" {
  project = %[1]q
  name    = "default"
}

resource "google_compute_network_peering" "foo" {
  name         = %[3]q
  network      = data.google_compute_network.foo.id
  peer_network = aiven_gcp_vpc_peering_connection.foo.self_link
}
`, c.peerProject, c.region, "test-tf-acc-"+acctest.RandString(7))
	var peeringID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             checkGCPPeeringDestroy,
		ExternalProviders: map[string]resource.ExternalProvider{
			"google": {Source: "hashicorp/google", VersionConstraint: ">=4.0.0,<5.0.0"},
		},
		Steps: []resource.TestStep{
			{
				Config: peeringConfig,
				Check: resource.ComposeTestCheckFunc(gcpPeeringChecks(c),
					resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
					resource.TestCheckResourceAttr("data."+peeringResource, "state", "PENDING_PEER"),
				),
			},
			{
				Config: fullConfig,
				Check: resource.ComposeTestCheckFunc(gcpPeeringChecks(c),
					resource.TestCheckResourceAttr("google_compute_network_peering.foo", "state", "ACTIVE"),
					func(state *terraform.State) error {
						peeringID = state.RootModule().Resources[peeringResource].Primary.ID
						return waitGCPPeeringActive(t.Context(), peeringID)
					},
				),
			},
			{
				Config:           fullConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check: resource.ComposeTestCheckFunc(gcpPeeringChecks(c),
					resource.TestCheckResourceAttr(peeringResource, "state", "ACTIVE"),
					resource.TestCheckResourceAttr("data."+peeringResource, "state", "ACTIVE"),
				),
			},
			{ResourceName: peeringResource, ImportState: true, ImportStateVerify: true},
			{
				// Keep the parent VPC so its deletion cannot hide a broken peering Delete.
				Config: gcpFixtureConfig(c),
				Check:  func(_ *terraform.State) error { return checkGCPPeeringDeleted(t.Context(), peeringID) },
			},
		},
	})
}

func TestAccAivenGCPPeeringConnection_replacement(t *testing.T) {
	c := getGCPConfig(t)
	config := gcpPeeringReplacementConfig(c, false)
	checkPeeringExists := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
		func(state *terraform.State) error {
			id, err := pluginvpc.ParseProjectPeeringID(state.RootModule().Resources[peeringResource].Primary.ID)
			if err != nil {
				return err
			}
			client, err := acc.GetTestGenAivenClient()
			if err != nil {
				return err
			}
			// Read the parent VPC directly: state can retain the replacement after
			// deleting the deposed instance has removed their shared connection.
			response, err := client.VpcGet(t.Context(), id.Project, id.ProjectVPCID)
			if err != nil {
				return err
			}
			for _, connection := range response.PeeringConnections {
				if connection.PeerCloudAccount == id.PeerCloudAccount && connection.PeerVpc == id.PeerVPC {
					if connection.State != vpc.VpcPeeringConnectionStateTypePendingPeer {
						return fmt.Errorf("GCP peering %q: expected PENDING_PEER in the API, got %s", id, connection.State)
					}
					return nil
				}
			}
			return fmt.Errorf("GCP peering %q is present in Terraform state but missing from VpcGet after apply", id)
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             checkGCPPeeringDestroy,
		Steps: []resource.TestStep{
			{Config: config, Check: checkPeeringExists},
			{
				Config: config,
				Taint:  []string{peeringResource},
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionDestroyBeforeCreate),
					plancheck.ExpectResourceAction("aiven_project_vpc.project_vpc", plancheck.ResourceActionNoop),
				}},
				Check: checkPeeringExists,
			},
			{
				Config: gcpPeeringReplacementConfig(c, true),
				Taint:  []string{peeringResource},
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionCreateBeforeDestroy),
					plancheck.ExpectResourceAction("aiven_project_vpc.project_vpc", plancheck.ResourceActionNoop),
				}},
				ExpectError: regexp.MustCompile(`already exists in state PENDING_PEER`),
			},
			{
				// A failed apply skips Check. Refresh separately to verify the
				// original connection remains managed and present in the API.
				RefreshState:       true,
				ExpectNonEmptyPlan: true, // The original resource is still tainted.
				Check:              checkPeeringExists,
			},
		},
	})
}

func TestAccAivenGCPPeeringConnection_ignoreVPCIDChanges(t *testing.T) {
	c := getGCPConfig(t)
	checks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(peeringResource, "vpc_id", "aiven_project_vpc.project_vpc", "id"),
		resource.TestCheckResourceAttrPair(peeringResource, "project", "aiven_project_vpc.project_vpc", "project"),
		resource.TestCheckResourceAttrPair(peeringResource, "project_vpc_id", "aiven_project_vpc.project_vpc", "project_vpc_id"),
		resource.TestCheckResourceAttr(peeringResource, "state", "PENDING_PEER"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             checkGCPPeeringDestroy,
		Steps: []resource.TestStep{
			{
				Config: gcpPeeringIgnoreVPCIDChangesConfig(c, "aiven_project_vpc.project_vpc.id"),
				Check:  checks,
			},
			{
				// A placeholder VPC ID is enough: ignore_changes must keep the original,
				// and the plan check stops any unexpected replacement before apply.
				Config: gcpPeeringIgnoreVPCIDChangesConfig(c, `"${aiven_project_vpc.project_vpc.project}/00000000-0000-4000-8000-000000000000"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: checks,
			},
			{
				Config: gcpPeeringIgnoreProjectVPCChangesConfig(c, "aiven_project_vpc.project_vpc.project", "aiven_project_vpc.project_vpc.project_vpc_id"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: checks,
			},
			{
				Config: gcpPeeringIgnoreProjectVPCChangesConfig(c, `"ignored-project"`, `"00000000-0000-4000-8000-000000000000"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: checks,
			},
		},
	})
}

func TestAccAivenGCPPeeringConnection_backwardCompat(t *testing.T) {
	c := getGCPConfig(t)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		CheckDestroy: checkGCPPeeringDestroy,
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           gcpPeeringConfig(c),
			OldProviderVersion: "4.62.0",
			Checks:             gcpPeeringChecks(c),
		}),
	})
}

func gcpPeeringChecks(c gcpConfig) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrPair(peeringResource, "vpc_id", "aiven_project_vpc.project_vpc", "id"),
		resource.TestCheckResourceAttr(peeringResource, "gcp_project_id", c.peerProject),
		resource.TestCheckResourceAttr(peeringResource, "peer_vpc", "default"),
		resource.TestCheckResourceAttrSet(peeringResource, "self_link"),
		resource.TestCheckResourceAttrSet(peeringResource, "state_info.to_project_id"),
		resource.TestCheckResourceAttrSet(peeringResource, "state_info.to_vpc_network"),
		resource.TestMatchResourceAttr(peeringResource, "state", regexp.MustCompile(`^(ACTIVE|PENDING_PEER)$`)),
		resource.TestCheckResourceAttrPair("data."+peeringResource, "id", peeringResource, "id"),
		resource.TestCheckResourceAttrPair("data."+peeringResource, "self_link", peeringResource, "self_link"),
		func(state *terraform.State) error {
			r := state.RootModule().Resources[peeringResource]
			want := fmt.Sprintf("%s/%s/default", r.Primary.Attributes["vpc_id"], c.peerProject)
			if r.Primary.ID != want {
				return fmt.Errorf("expected legacy peering ID %q, got %q", want, r.Primary.ID)
			}
			return nil
		},
	)
}

func waitGCPPeeringActive(ctx context.Context, value string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	id, err := pluginvpc.ParseProjectPeeringID(value)
	if err != nil {
		return err
	}
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		return err
	}
	return retry.Do(func() error {
		connection, err := id.Find(ctx, client)
		if err != nil {
			return retry.Unrecoverable(err)
		}
		if connection.State != vpc.VpcPeeringConnectionStateTypeActive {
			return fmt.Errorf("waiting for ACTIVE peering, got %s", connection.State)
		}
		return nil
	}, retry.Context(ctx), retry.Attempts(0), retry.Delay(5*time.Second), retry.DelayType(retry.FixedDelay), retry.LastErrorOnly(true))
}

func checkGCPPeeringDeleted(ctx context.Context, value string) error {
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
		return fmt.Errorf("GCP peering %q still exists in state %s", value, connection.State)
	}
	return nil
}

func checkGCPPeeringDestroy(state *terraform.State) error {
	for _, r := range state.RootModule().Resources {
		if r.Type == "aiven_gcp_vpc_peering_connection" {
			if err := checkGCPPeeringDeleted(context.Background(), r.Primary.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func gcpFixtureConfig(c gcpConfig) string {
	return fmt.Sprintf(`
resource "aiven_project_vpc" "project_vpc" {
  project      = %[1]q
  cloud_name   = "google-%[2]s"
  network_cidr = "10.0.0.0/24"
}
`, c.project, c.region)
}

func gcpPeeringConfig(c gcpConfig) string {
	return gcpFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_gcp_vpc_peering_connection" "foo" {
  vpc_id         = aiven_project_vpc.project_vpc.id
  gcp_project_id = %[1]q
  peer_vpc       = "default"
}

data "aiven_gcp_vpc_peering_connection" "foo" {
  vpc_id         = aiven_gcp_vpc_peering_connection.foo.vpc_id
  gcp_project_id = aiven_gcp_vpc_peering_connection.foo.gcp_project_id
  peer_vpc       = aiven_gcp_vpc_peering_connection.foo.peer_vpc
}
`, c.peerProject)
}

func gcpPeeringReplacementConfig(c gcpConfig, createBeforeDestroy bool) string {
	return gcpFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_gcp_vpc_peering_connection" "foo" {
  vpc_id         = aiven_project_vpc.project_vpc.id
  gcp_project_id = %[1]q
  peer_vpc       = "default"

  lifecycle {
    create_before_destroy = %[2]t
  }
}
`, c.peerProject, createBeforeDestroy)
}

func gcpPeeringIgnoreVPCIDChangesConfig(c gcpConfig, vpcIDExpression string) string {
	return gcpFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_gcp_vpc_peering_connection" "foo" {
  vpc_id         = %[2]s
  gcp_project_id = %[1]q
  peer_vpc       = "default"

  lifecycle {
    ignore_changes = [vpc_id]
  }
}
`, c.peerProject, vpcIDExpression)
}

func gcpPeeringIgnoreProjectVPCChangesConfig(c gcpConfig, projectExpression, vpcIDExpression string) string {
	return gcpFixtureConfig(c) + fmt.Sprintf(`
resource "aiven_gcp_vpc_peering_connection" "foo" {
  project        = %[2]s
  project_vpc_id = %[3]s
  gcp_project_id = %[1]q
  peer_vpc       = "default"

  lifecycle {
    ignore_changes = [project, project_vpc_id]
  }
}
`, c.peerProject, projectExpression, vpcIDExpression)
}
