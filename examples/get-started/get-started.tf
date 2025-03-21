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
  name = "ORGANIZATION_NAME" # Replace ORGANIZATION_NAME with your organization's name
}

# Your organization's users
data "aiven_organization_user_list" "users" {
  id = data.aiven_organization.main.id
}

# Grant a user organization admin permissions
resource "aiven_organization_permission" "example_org_permissions" {
  organization_id = data.aiven_organization.main.id
  resource_id     = data.aiven_organization.main.id
  resource_type   = "organization"

  permissions {
    permissions = [
      "role:organization:admin"
    ]
    principal_id   = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].user_email == "USER_EMAIL"]) # Replace USER_EMAIL with the user's email address
    principal_type = "user"
  }
}

# Create a project in your organization 
resource "aiven_project" "example_project" {
  project   = "ORGANIZATION_NAME-first-project" # Replace ORGANIZATION_NAME with your organization's name
  parent_id = data.aiven_organization.main.id
}

# Create a user group 
resource "aiven_organization_user_group" "developers" {
  organization_id = data.aiven_organization.main.id
  name            = "Example user group"
  description     = "The first user group for this organization."
}

# Add an existing organization user to the group
resource "aiven_organization_user_group_member" "developers" {
  group_id        = aiven_organization_user_group.developers.group_id
  organization_id = data.aiven_organization.main.id
  user_id         = one([for user in data.aiven_organization_user_list.users.users : user.user_id if user.user_info[0].user_email == "USER_EMAIL"]) # Replace USER_EMAIL with the user's email address
}

# Grant the group the developer role and
# access to create project integrations
resource "aiven_organization_permission" "project_developers" {
  organization_id = data.aiven_organization.main.id
  resource_id     = aiven_project.example_project.project
  resource_type   = "project"

  permissions {
    permissions = [
      "project:integrations:write",
      "developer"
    ]
    principal_id   = aiven_organization_user_group.developers.group_id
    principal_type = "user_group"
  }
}