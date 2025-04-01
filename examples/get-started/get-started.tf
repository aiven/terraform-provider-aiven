terraform {
  required_version = ">=0.13"
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.0.0, <5.0.0"
    }
  }
}

variable "aiven_token" {}


provider "aiven" {
  api_token = var.aiven_token
}

# Your organization
data "aiven_organization" "main" {
  name = "ORGANIZATION_NAME"
}

# List of users in your organization
data "aiven_organization_user_list" "users" {
  name = "ORGANIZATION_NAME"
}

# Create a project in your organization 
resource "aiven_project" "example_project" {
  project    = "ORGANIZATION_NAME-first-project"
  parent_id = data.aiven_organization.main.id
}

# Create a user group 
resource "aiven_organization_user_group" "example_group" {
  organization_id = data.aiven_organization.main.id
  name       = "Example user group"
  description = "The first user group for this organization."
}

# Add an existing organization user to the group
resource "aiven_organization_user_group_member" "group-members" {
  group_id      = aiven_organization_user_group.example_group.group_id 
  organization_id = data.aiven_organization.main.id
  user_id = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].user_email == "EMAIL_ADDRESS"])
}

# Give the group access to your project with the developer role
resource "aiven_organization_group_project" "group-proj" {
  group_id      = aiven_organization_user_group.group.group_id
  project = aiven_project.example_project.project
  role    = "developer"
}

