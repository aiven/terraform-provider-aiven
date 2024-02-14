variable "aiven_api_token" {}
variable "prod_project_name" {}
variable "qa_project_name" {}
variable "dev_project_name" {}

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


# Create organization
resource "aiven_organization" "org" {
  name = "Example Organization"
}


# Create units within organization
resource "aiven_organizational_unit" "unit-eng" {
  name = "Engineering"
  parent_id = aiven_organization.org.id
}

resource "aiven_organizational_unit" "unit-fin" {
  name = "Finance"
  parent_id = aiven_organization.org.id
}

# Create projects in units

# Engineering projects
resource "aiven_project" "prod-eng" {
  project    = "${var.prod_project_name}-eng"
  parent_id = aiven_organizational_unit.unit-eng.id
}

resource "aiven_project" "qa-eng" {
  project    = "${var.qa_project_name}-eng"
  parent_id = aiven_organizational_unit.unit-eng.id
}

resource "aiven_project" "dev-eng" {
  project    = "${var.dev_project_name}-eng"
  parent_id = aiven_organizational_unit.unit-eng.id
}

# Finance projects
resource "aiven_project" "prod-fin" {
  project    = "${var.prod_project_name}-fin"
  parent_id = aiven_organizational_unit.unit-fin.id
}

resource "aiven_project" "qa-fin" {
  project    = "${var.qa_project_name}-fin"
  parent_id = aiven_organizational_unit.unit-fin.id
}

resource "aiven_project" "dev-fin" {
  project    = "${var.dev_project_name}-fin"
  parent_id = aiven_organizational_unit.unit-fin.id
}