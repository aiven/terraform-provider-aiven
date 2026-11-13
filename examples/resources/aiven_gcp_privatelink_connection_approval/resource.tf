terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

provider "google" {
  project = "my-google-project"
  region  = "europe-west1"
}

# The existing Aiven service must be in an Aiven VPC in google-europe-west1,
# with privatelink_access enabled in its service user configuration.
resource "aiven_gcp_privatelink" "example" {
  project      = "my-aiven-project"
  service_name = "my-service"
}

resource "google_compute_network" "consumer" {
  name                    = "psc-consumer"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "consumer" {
  name          = "psc-consumer"
  network       = google_compute_network.consumer.id
  region        = "europe-west1"
  ip_cidr_range = "10.80.0.0/24"
}

resource "google_compute_address" "endpoint" {
  name         = "aiven-psc"
  address_type = "INTERNAL"
  subnetwork   = google_compute_subnetwork.consumer.id
  region       = "europe-west1"
}

resource "google_compute_forwarding_rule" "endpoint" {
  name                  = "aiven-psc"
  region                = "europe-west1"
  network               = google_compute_network.consumer.id
  ip_address            = google_compute_address.endpoint.id
  target                = aiven_gcp_privatelink.example.google_service_attachment
  load_balancing_scheme = ""
}

resource "aiven_gcp_privatelink_connection_approval" "example" {
  project           = aiven_gcp_privatelink.example.project
  service_name      = aiven_gcp_privatelink.example.service_name
  psc_connection_id = google_compute_forwarding_rule.endpoint.psc_connection_id
  user_ip_address   = google_compute_forwarding_rule.endpoint.ip_address
}
