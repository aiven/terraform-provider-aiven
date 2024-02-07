terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

variable "aiven_api_token" {}


provider "aiven" {
  api_token = var.aiven_api_token
}

# Your organization
data "aiven_organization" "org" {
  name = "YOUR_ORGANIZATION_NAME"
}

# Create a project in your organization 
resource "aiven_project" "example-project" {
  project    = "ORGANIZATION_NAME-first-project"
  parent_id = data.aiven_organization.org.id
}

# Create a user group 
resource "aiven_organization_user_group" "group" {
  organization_id = data.aiven_organization.org.id
  name       = "Example user group"
  description = "The first user group for this organization."
}

# Add an existing organization user to the group
resource "aiven_organization_user_group_member" "group-members" {
  group_id      = aiven_organization_user_group.group.group_id 
  organization_id = data.aiven_organization.org.id
  user_id = "USER_ID"
}

# Give the group access to your project with the developer role
resource "aiven_organization_group_project" "group-proj" {
  group_id      = aiven_organization_user_group.group.group_id
  project = aiven_project.example-project.project
  role    = "developer"
}

