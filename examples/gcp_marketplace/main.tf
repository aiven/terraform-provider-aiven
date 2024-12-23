terraform {
  required_version = ">=0.13"
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}

# Creating a service in an existing project

data "aiven_project" "existing_project" {
  project = "existing_project"
}

resource "aiven_kafka" "kafka" {
  project      = data.aiven_project.existing_project.id
  cloud_name   = "google-europe-north1"
  plan         = "gcp-marketplace-startup-2"
  service_name = "kafka"
}

# Creating a service in a new project

data "aiven_organization" "org" {
  name = "YOUR_ORGANIZATION_NAME"
}

resource "aiven_project" "new_project" {
  project    = "new-project"
  parent_id = data.aiven_organization.org.id # This is required for new marketplace projects
}

resource "aiven_opensearch" "opensearch" {
  project      = data.aiven_project.existing_project.id
  cloud_name   = "google-europe-north1"
  plan         = "gcp-marketplace-startup-4"
  service_name = "opensearch"
}
