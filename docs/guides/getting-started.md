---
parent: Guides
page_title: "Getting Started"
sidebar_current: "docs-aiven-getting-started"
description: |-
  The Aiven Terraform Provider is a way to access all of your Aiven services within Terraform. Use this provider to set up and teardown services and test out configurations. Any issues, please email support@aiven.io
---

## Installing the Aiven Provider

```hcl
terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = "2.X.X"
    }
  }
}
```

#### TIP: Using with Terraform Cloud

Terraform Cloud will also need the provider. In this case, it is best to install it locally and make sure you copy it to `linux_amd64` inside the plugins directory.

## Set up

You will need an API token from your account, which you can get through the [Aiven Web Console](https://console.aiven.io/profile/auth) or the [Aiven CLI](https://github.com/aiven/aiven-cli)

## Your first run

`$ terraform init`
`$ terraform version`

You should now see terraform and terraform-provider-aiven with the version number as output.

### Hello World

It is difficult to make an Open Source Database Management Service say Hello World, so maybe we will just setup a project and deploy Postgres.

```hcl
terraform {
  required_providers {
    aiven = "1.2.4"
  }
}

variable "aiven_api_token" {
  type = string
}

provider "aiven" {
  api_token = var.aiven_api_token
}

resource "aiven_project" "hello" {
  project = "hello-project"
}

resource "aiven_service" "world" {
	project = "${aiven_project.hello.project}"
	cloud_name = "google-europe-west1"
	plan = "business-4"
	service_name = "world-pg"
	service_type = "pg"
	maintenance_window_dow = "monday"
	maintenance_window_time = "12:00:00"
	pg_user_config {
		pg {
			idle_in_transaction_session_timeout = 900
		}
		pg_version = "10"
	}
}

```

Save this as `getting-started.tf`

`$ terraform plan # This will show you what Terraform thinks you want to do and you can confirm it before applying`

`$ terraform apply # Now we really do it`

`$ terraform destroy # When you want to sound cool and undo all the things in the script`
