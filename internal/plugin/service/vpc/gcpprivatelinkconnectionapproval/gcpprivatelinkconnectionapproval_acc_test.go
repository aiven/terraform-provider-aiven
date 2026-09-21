package gcpprivatelinkconnectionapproval_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aiven/go-client-codegen/handler/privatelink"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

const approvalResource = "aiven_gcp_privatelink_connection_approval.test"

func TestAccAivenGCPPrivatelinkConnectionApproval(t *testing.T) {
	acc.SkipIfNotAcc(t)
	fixture := approvalFixture(t)
	// The parent owns the infrastructure and destroys it after both subtests.
	// Subtests manage only approval, so switching providers cannot recreate Kafka.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders:        googleProvider(),
		Steps: []resource.TestStep{
			{
				Config: fixture,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("google_compute_forwarding_rule.endpoint", "psc_connection_id"),
					resource.TestCheckResourceAttrSet("google_compute_address.endpoint", "address"),
					func(s *terraform.State) error {
						runApprovalTests(t, s)
						return nil
					},
				),
			},
		},
	})
}

func runApprovalTests(t *testing.T, s *terraform.State) {
	t.Helper()

	pl := s.RootModule().Resources["aiven_gcp_privatelink.test"].Primary.Attributes
	project, service := pl["project"], pl["service_name"]
	pscID := s.RootModule().Resources["google_compute_forwarding_rule.endpoint"].Primary.Attributes["psc_connection_id"]
	ip := s.RootModule().Resources["google_compute_address.endpoint"].Primary.Attributes["address"]
	config := approvalConfig(project, service, pscID, ip)
	checks := approvalChecks(project, service, pscID, ip)

	// Both subtests use the same endpoint and must run sequentially.
	t.Run("basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: config, Check: resource.ComposeTestCheckFunc(checks, resource.TestCheckResourceAttr(approvalResource, "state", "active"))},
				{ResourceName: approvalResource, ImportState: true, ImportStateVerify: true},
				{
					ResourceName: approvalResource, ImportState: true, ImportStateVerify: true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						r := s.RootModule().Resources[approvalResource]
						return r.Primary.ID + "/" + r.Primary.Attributes["psc_connection_id"], nil
					},
				},
				{
					Config:   approvalConfig(project, service, pscID, "10.20.0.250"),
					PlanOnly: true,
					// Terraform can wrap the diagnostic between any two words.
					ExpectError: regexp.MustCompile(`IP\s+address\s+of\s+an\s+approved\s+Google\s+Private\s+Service\s+Connect\s+connection\s+cannot\s+be\s+changed`),
				},
				{
					// Destroying the Terraform approval must leave the endpoint active.
					Config:  config,
					Destroy: true,
					Check:   approvalStillExists(t.Context(), project, service, pscID),
				},
			},
		})
	})

	// The SDK writes its own state even if basic has already approved the endpoint.
	t.Run("backwardCompat", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
				TFConfig:           config,
				OldProviderVersion: "4.62.0",
				Checks:             checks,
			}),
		})
	})
}

func approvalChecks(project, service, pscID, ip string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(approvalResource, "project", project),
		resource.TestCheckResourceAttr(approvalResource, "service_name", service),
		resource.TestCheckResourceAttr(approvalResource, "id", project+"/"+service),
		resource.TestCheckResourceAttr(approvalResource, "psc_connection_id", pscID),
		resource.TestCheckResourceAttr(approvalResource, "user_ip_address", ip),
		resource.TestCheckResourceAttrSet(approvalResource, "privatelink_connection_id"),
	)
}

func approvalStillExists(ctx context.Context, project, service, pscID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := acc.GetTestGenAivenClient()
		if err != nil {
			return err
		}
		connections, err := client.ServicePrivatelinkGoogleConnectionList(ctx, project, service)
		if err != nil {
			return err
		}
		for _, c := range connections {
			if c.PscConnectionId == pscID && c.State == privatelink.ConnectionStateTypeActive {
				return nil
			}
		}
		return fmt.Errorf("approved connection %q is no longer active", pscID)
	}
}

func googleProvider() map[string]resource.ExternalProvider {
	return map[string]resource.ExternalProvider{
		"google": {Source: "hashicorp/google", VersionConstraint: "=6.15.0"},
	}
}

func approvalFixture(t *testing.T) string {
	t.Helper()
	env := acc.RequireEnvVars(t, "GCP_PRIVATE_LINK_TEST", "GOOGLE_PROJECT")
	name := "test-tf-psc-" + acctest.RandString(7)
	return fmt.Sprintf(`
provider "google" {
  project = %[1]q
  region  = "europe-west1"
}

resource "aiven_project_vpc" "test" {
  project      = %[2]q
  cloud_name   = "google-europe-west1"
  network_cidr = "10.0.1.0/24"
}

resource "aiven_kafka" "test" {
  project        = aiven_project_vpc.test.project
  project_vpc_id = aiven_project_vpc.test.id
  service_name   = %[3]q
  cloud_name     = "google-europe-west1"
  plan           = "startup-4"

  kafka_user_config {
    privatelink_access {
      kafka = true
    }
  }
}

resource "aiven_gcp_privatelink" "test" {
  project      = aiven_kafka.test.project
  service_name = aiven_kafka.test.service_name
}

resource "google_compute_network" "consumer" {
  name                    = %[3]q
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "consumer" {
  name          = %[3]q
  ip_cidr_range = "10.20.0.0/24"
  network       = google_compute_network.consumer.id
}

resource "google_compute_address" "endpoint" {
  name         = %[3]q
  subnetwork   = google_compute_subnetwork.consumer.id
  address_type = "INTERNAL"
}

resource "google_compute_forwarding_rule" "endpoint" {
  name                  = %[3]q
  load_balancing_scheme = ""
  target                = aiven_gcp_privatelink.test.google_service_attachment
  network               = google_compute_network.consumer.id
  ip_address            = google_compute_address.endpoint.id
}
`, env["GOOGLE_PROJECT"], acc.ProjectName(), name)
}

func approvalConfig(project, service, pscID, ip string) string {
	return fmt.Sprintf(`
resource "aiven_gcp_privatelink_connection_approval" "test" {
  project           = %q
  service_name      = %q
  psc_connection_id = %q
  user_ip_address   = %q
}
`, project, service, pscID, ip)
}
