package awsvpcpeeringconnection_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	pluginvpc "github.com/aiven/terraform-provider-aiven/internal/plugin/vpc"
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

func TestAccAivenAWSVPCPeeringConnection_basic(t *testing.T) {
	c := getAWSConfig(t)
	config := awsPeeringConfig(c, c.region)
	emptyRegionConfig := awsPeeringConfig(c, "")
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
			{
				Config:           emptyRegionConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionUpdate)}},
				Check: resource.ComposeTestCheckFunc(awsPeeringChecks(c),
					resource.TestCheckResourceAttr(peeringResource, "aws_vpc_region", ""),
					func(state *terraform.State) error {
						if state.RootModule().Resources[peeringResource].Primary.ID != peeringID {
							return fmt.Errorf("clearing the region changed the peering ID")
						}
						return nil
					},
				),
			},
			{Config: emptyRegionConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(peeringResource, plancheck.ResourceActionUpdate)}}},
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
	steps := acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
		TFConfig: awsPeeringConfig(c, c.region), OldProviderVersion: "4.62.0", Checks: awsPeeringChecks(c),
	})
	for i := range steps {
		if steps[i].ExternalProviders == nil {
			steps[i].ExternalProviders = make(map[string]resource.ExternalProvider)
		}
		steps[i].ExternalProviders["aws"] = awsProvider
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acc.TestAccPreCheck(t) }, CheckDestroy: checkAWSPeeringDestroy, Steps: steps,
	})
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
				return fmt.Errorf("expected legacy peering ID %q, got %q", want, r.Primary.ID)
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
