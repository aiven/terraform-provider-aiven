package gcpprivatelink_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenGCPPrivatelink_basic(t *testing.T) {
	skipGCPPrivatelinkTests(t)
	gcpProject := acc.RequireEnvVars(t, "GOOGLE_PROJECT")["GOOGLE_PROJECT"]

	resourceName := "aiven_gcp_privatelink.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	config := testAccGCPPrivatelinkResource(rName) + fmt.Sprintf(`
provider "google" {
  project = %q
  region  = "europe-west1"
}
`, gcpProject)
	approvalConfig := config + testAccGCPPrivatelinkEndpoint(rName)
	const approvalName = "aiven_gcp_privatelink_connection_approval.endpoint"
	approvalChecks := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(approvalName, "privatelink_connection_id"),
		resource.TestCheckResourceAttrPair(approvalName, "psc_connection_id", "google_compute_forwarding_rule.endpoint", "psc_connection_id"),
		resource.TestCheckResourceAttrPair(approvalName, "user_ip_address", "google_compute_address.endpoint", "address"),
		resource.TestMatchResourceAttr(approvalName, "state", regexp.MustCompile("^(user-approved|connected|active)$")),
	)
	var identityImport acc.IdentityImportVersionCheck
	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{&identityImport},
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"google": {Source: "hashicorp/google", VersionConstraint: "=6.15.0"},
		},
		CheckDestroy: testAccCheckAivenGCPPrivatelinkResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenGCPPrivatelinkAttributes("data.aiven_gcp_privatelink.pr"),
					resource.TestCheckResourceAttr(resourceName, "state", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "google_service_attachment"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: resourceName, ImportState: true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity, SkipFunc: identityImport.Skip,
			},
			// Reuse the service and VPC for a real consumer endpoint and approval.
			{
				Config: approvalConfig,
				Check: resource.ComposeTestCheckFunc(
					approvalChecks,
					func(state *terraform.State) error {
						// The SDK returns at user-approved. Wait for active, then let
						// the next step refresh Terraform state before verifying imports.
						attrs := state.RootModule().Resources[approvalName].Primary.Attributes
						connectionID := attrs["privatelink_connection_id"]
						client := acc.GetTestAivenClient().GCPPrivatelink
						ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
						defer cancel()
						wait := &retry.StateChangeConf{
							Pending: []string{"pending-user-approval", "user-approved", "connected"},
							Target:  []string{"active"},
							Refresh: func() (any, string, error) {
								if err := client.Refresh(ctx, attrs["project"], attrs["service_name"]); err != nil {
									return nil, "", err
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
				Config:           approvalConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check: resource.ComposeTestCheckFunc(
					approvalChecks,
					resource.TestCheckResourceAttr(approvalName, "state", "active"),
				),
			},
			{ResourceName: approvalName, ImportState: true, ImportStateVerify: true},
			{
				ResourceName: approvalName, ImportState: true, ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					attrs := state.RootModule().Resources[approvalName].Primary.Attributes
					return schemautil.BuildResourceID(attrs["project"], attrs["service_name"], attrs["psc_connection_id"]), nil
				},
			},
			// Approval deletion is state-only; the consumer owns the connection.
			{Config: config, Check: resource.TestCheckResourceAttr(resourceName, "state", "active")},
		},
	})
}

func TestAccAivenGCPPrivatelink_backwardCompat(t *testing.T) {
	skipGCPPrivatelinkTests(t)

	resourceName := "aiven_gcp_privatelink.foo"
	dataSourceName := "data.aiven_gcp_privatelink.pr"
	projectName := acc.ProjectName()
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	config := testAccGCPPrivatelinkResource(rName)

	var identityImport acc.IdentityImportVersionCheck
	testCase := resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{&identityImport},
		PreCheck:               func() { acc.TestAccPreCheck(t) },
		CheckDestroy:           testAccCheckAivenGCPPrivatelinkResourceDestroy,
		Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
			TFConfig:           config,
			OldProviderVersion: "4.60.0",
			Checks: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "project", projectName),
				resource.TestCheckResourceAttr(resourceName, "service_name", "test-acc-sr-"+rName),
				resource.TestCheckResourceAttr(resourceName, "state", "active"),
				resource.TestCheckResourceAttrSet(resourceName, "google_service_attachment"),
				resource.TestCheckResourceAttrPair(dataSourceName, "project", resourceName, "project"),
				resource.TestCheckResourceAttrPair(dataSourceName, "service_name", resourceName, "service_name"),
				resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
				resource.TestCheckResourceAttrPair(dataSourceName, "google_service_attachment", resourceName, "google_service_attachment"),
			),
		}),
	}
	testCase.Steps[1].ConfigPlanChecks.PreApply = []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}
	testCase.Steps = append(testCase.Steps, acc.ProviderRollbackSteps(t, testCase.Steps[0], testCase.Steps[1], &identityImport)...)
	resource.Test(t, testCase)
}

func skipGCPPrivatelinkTests(t *testing.T) {
	t.Helper()

	if _, ok := os.LookupEnv("GCP_PRIVATE_LINK_TEST"); !ok {
		t.Skip("environment variable GCP_PRIVATE_LINK_TEST not set")
	}
}

func testAccCheckAivenGCPPrivatelinkResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_gcp_privatelink" {
			continue
		}

		project, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		privatelink, err := c.GCPPrivatelink.Get(ctx, project, serviceName)
		var apiErr aiven.Error
		if common.IsCritical(err) && errors.As(err, &apiErr) && apiErr.Status != 500 {
			return fmt.Errorf("error getting a GCP Privatelink: %w", err)
		}

		if privatelink != nil {
			return fmt.Errorf("GCP Privatelink (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccGCPPrivatelinkResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = %q
}

resource "aiven_project_vpc" "aiven_vpc" {
  project      = data.aiven_project.foo.project
  cloud_name   = "google-europe-west1"
  network_cidr = "10.0.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aiven_kafka" "bar" {
  project        = data.aiven_project.foo.project
  cloud_name     = "google-europe-west1"
  plan           = "business-4"
  service_name   = "test-acc-sr-%s"
  project_vpc_id = aiven_project_vpc.aiven_vpc.id

  kafka_user_config {
    privatelink_access {
      kafka         = true
      kafka_connect = true
      kafka_rest    = true
    }
  }
}

resource "aiven_gcp_privatelink" "foo" {
  project      = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
}

data "aiven_gcp_privatelink" "pr" {
  project      = aiven_gcp_privatelink.foo.project
  service_name = aiven_gcp_privatelink.foo.service_name
}
`, acc.ProjectName(), name)
}

func testAccCheckAivenGCPPrivatelinkAttributes(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributes := s.RootModule().Resources[name].Primary.Attributes

		for _, attribute := range []string{"project", "service_name", "state", "google_service_attachment"} {
			if attributes[attribute] == "" {
				return fmt.Errorf("expected %q to be set", attribute)
			}
		}

		return nil
	}
}

// Consumer endpoint pattern: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_forwarding_rule#example-usage---forwarding-rule-vpc-psc
func testAccGCPPrivatelinkEndpoint(name string) string {
	return fmt.Sprintf(`
resource "google_compute_network" "consumer" {
  name                    = "test-acc-psc-%[1]s"
  auto_create_subnetworks = false
}
resource "google_compute_subnetwork" "consumer" {
  name          = "test-acc-psc-%[1]s"
  network       = google_compute_network.consumer.id
  region        = "europe-west1"
  ip_cidr_range = "10.80.0.0/24"
}
resource "google_compute_address" "endpoint" {
  name         = "test-acc-psc-%[1]s"
  address_type = "INTERNAL"
  subnetwork   = google_compute_subnetwork.consumer.id
  region       = "europe-west1"
}
resource "google_compute_forwarding_rule" "endpoint" {
  name                  = "test-acc-psc-%[1]s"
  region                = "europe-west1"
  network               = google_compute_network.consumer.id
  ip_address            = google_compute_address.endpoint.id
  target                = aiven_gcp_privatelink.foo.google_service_attachment
  load_balancing_scheme = ""
}
resource "aiven_gcp_privatelink_connection_approval" "endpoint" {
  project           = aiven_gcp_privatelink.foo.project
  service_name      = aiven_gcp_privatelink.foo.service_name
  user_ip_address   = google_compute_address.endpoint.address
  psc_connection_id = google_compute_forwarding_rule.endpoint.psc_connection_id
}
`, name)
}
