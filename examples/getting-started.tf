terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = ">= 2.0.0, < 3.0.0"
    }
  }
}

variable "aiven_api_token" {
  type = string
}

provider "aiven" {
  api_token = var.aiven_api_token
}

resource "aiven_project" "foo" {
  project = "project-1"
  account_id = aiven_account_team.foo.account_id
}

# Account
resource "aiven_account" "foo" {
  name = "account1"
}

# Account team
resource "aiven_account_team" "foo" {
  account_id = aiven_account.foo.account_id
  name = "account_team1"
}

# Account team project
resource "aiven_account_team_project" "foo" {
  account_id = aiven_account.foo.account_id
  team_id = aiven_account_team.foo.team_id
  project_name = aiven_project.foo.project
  team_type = "admin"
}

# Account team member
resource "aiven_account_team_member" "foo" {
  account_id = aiven_account.foo.account_id
  team_id = aiven_account_team.foo.team_id
  user_email = "user+1@example.com"
}

data "aiven_account_team" "team" {
  name = aiven_account_team.foo.name
  account_id = aiven_account_team.foo.account_id
}

data "aiven_account_team_member" "member" {
  team_id = aiven_account_team_member.foo.team_id
  account_id = aiven_account_team_member.foo.account_id
  user_email = aiven_account_team_member.foo.user_email
}

