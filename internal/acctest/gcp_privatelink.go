package acctest

import (
	"context"
	"errors"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// GCPPrivatelinkFixture creates an Aiven VPC, Kafka service and PrivateLink producer.
func GCPPrivatelinkFixture(name string) string {
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
`, ProjectName(), name)
}

// GCPPrivatelinkEndpointFixture creates a consumer PSC endpoint for GCPPrivatelinkFixture.
// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_forwarding_rule#example-usage---forwarding-rule-vpc-psc
func GCPPrivatelinkEndpointFixture(name, project string) string {
	return fmt.Sprintf(`
provider "google" {
  project = %[2]q
  region  = "europe-west1"
}

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
`, name, project)
}

// TestAccCheckGCPPrivatelinkDestroyed verifies cleanup of the shared PrivateLink infrastructure.
func TestAccCheckGCPPrivatelinkDestroyed(s *terraform.State) error {
	c := GetTestAivenClient()
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
