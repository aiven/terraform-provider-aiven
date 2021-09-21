variable "avn_api_token" {}
variable "prod_project_name" {}
variable "qa_project_name" {}
variable "dev_project_name" {}

terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = ">= 2.0.0, < 3.0.0"
    }
  }
}

################################################
################################################
#################### AIVEN #####################
################################################
################################################
# This is a simple block to add authentication 
# to the Aiven Terraform provider for the 
# underlying REST client.
################################################
provider "aiven" {
  api_token = var.avn_api_token
}


################################################
################################################
############# CORPORATE ACCOUNT ################
################################################
################################################
# Create an account entity to manage RBAC access
# controls across all of your projects, users,
# and SSO authentication integrations.
################################################
resource "aiven_account" "acct" {
  name = "demo-tech"
}


################################################
################################################
################## PROJECT #####################
################################################
################################################
# Projects are grouping of services, users, and
# privileges. Projects support multi-region, 
# and multi-cloud deployments. Most customers
# break up projects and services into:
# - production
# - QA
# - development
################################################
resource "aiven_project" "prj-prod" {
  project = var.prod_project_name
  account_id = aiven_account.acct.account_id
}
resource "aiven_project" "prj-qa" {
  project = var.qa_project_name
  account_id = aiven_account.acct.account_id
}
resource "aiven_project" "prj-dev" {
  project = var.dev_project_name
  account_id = aiven_account.acct.account_id
}

################################################
################################################
################### TEAMS ######################
################################################
################################################
# Teams are roles, or groups of users that have
# the same privileges across projects. Most
# customers break up teams into:
# - admins
# - operators
# - development
# - qa
# - unassinged: default role for SSO registration
# You can see more details about Aiven user 
# roles https://bit.ly/3garJcr
################################################
resource "aiven_account_team" "tm-admin" {
  account_id = aiven_account.acct.account_id
  name = "Admins"
}
resource "aiven_account_team" "tm-ops" {
  account_id = aiven_account.acct.account_id
  name = "Operations"
}
resource "aiven_account_team" "tm-dev" {
  account_id = aiven_account.acct.account_id
  name = "Developers"
}
resource "aiven_account_team" "tm-qa" {
  account_id = aiven_account.acct.account_id
  name = "Quality Assurance"
}
resource "aiven_account_team" "tm-default" {
  account_id = aiven_account.acct.account_id
  name = "Unassigned"
}

################################################
################################################
############ TEAM / PROJECT RBAC ###############
################################################
################################################
# Define team privileges in projects using the
# following roles:
# - admin: Billing + operator
# - operators: Service CRUD + developer
# - development: Connection creds + read_only
# - read_only: view service status
# You can see more details about Aiven user 
# roles https://bit.ly/3garJcr
################################################

# admin team
resource "aiven_account_team_project" "rbac-prod-admin" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-admin.team_id
  project_name = aiven_project.prj-prod.project
  team_type = "admin"
}
resource "aiven_account_team_project" "rbac-qa-admin" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-admin.team_id
  project_name = aiven_project.prj-qa.project
  team_type = "admin"
}
resource "aiven_account_team_project" "rbac-dev-admin" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-admin.team_id
  project_name = aiven_project.prj-dev.project
  team_type = "admin"
}

# operator team
resource "aiven_account_team_project" "rbac-prod-ops" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-ops.team_id
  project_name = aiven_project.prj-prod.project
  team_type = "operator"
}
resource "aiven_account_team_project" "rbac-qa-ops" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-ops.team_id
  project_name = aiven_project.prj-qa.project
  team_type = "operator"
}
resource "aiven_account_team_project" "rbac-dev-ops" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-ops.team_id
  project_name = aiven_project.prj-dev.project
  team_type = "operator"
}

# developer team
resource "aiven_account_team_project" "rbac-prod-dev" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-dev.team_id
  project_name = aiven_project.prj-prod.project
  team_type = "read_only"
}
resource "aiven_account_team_project" "rbac-qa-dev" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-dev.team_id
  project_name = aiven_project.prj-qa.project
  team_type = "developer"
}
resource "aiven_account_team_project" "rbac-dev-dev" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-dev.team_id
  project_name = aiven_project.prj-dev.project
  team_type = "developer"
}

# qa team
resource "aiven_account_team_project" "rbac-prod-qa" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-qa.team_id
  project_name = aiven_project.prj-prod.project
  team_type = "read_only"
}
resource "aiven_account_team_project" "rbac-qa-qa" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-qa.team_id
  project_name = aiven_project.prj-qa.project
  team_type = "read_only"
}
resource "aiven_account_team_project" "rbac-dev-qa" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-qa.team_id
  project_name = aiven_project.prj-dev.project
  team_type = "read_only"
}

################################################
################################################
################### USERS ######################
################################################
################################################
# You can define explict team membership using
# email addresses.
################################################
resource "aiven_account_team_member" "u-david" {
  account_id = aiven_account.acct.account_id
  team_id = aiven_account_team.tm-admin.team_id
  user_email = "david@aiven.io"
}
