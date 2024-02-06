---
page_title: "Creating static IP addresses"
---

# Introduction

You can create static IP addresses for your services for an additional charge. Static IP addresses belong to a project and are created in the cloud you specify.

For more information, see the article on [static IP addresses in Aiven](https://aiven.io/docs/platform/concepts/static-ips).

# Example

The following example file creates 6 static IP addresses for a PostgreSQL service in the Google Cloud europe-west-1 region. The `static_ip` user configuration option is also set to `true` to enable static IP addresses for the service.

```hcl
terraform {
    required_providers {
        aiven = {
            source = "aiven/aiven"
        }
    }
}

variable "aiven_api_token" {
    type = string
}

variable "aiven_project_name" {
    type = string
}

provider "aiven" {
    api_token = var.aiven_api_token
}

resource "aiven_static_ip" "ips" {
    count = 6

    project = var.aivenk_project_name
    cloud_name = "google-europe-west-1"
}

resource "aiven_pg" "pg" {
    project = var.aiven_project_name
    cloud_name = "google-europe-west-1"
    plan = "startup-4"
    service_name = "my-service"

    static_ips = toset([
        aiven_static_ip.ips[0].static_ip_address_id,
        aiven_static_ip.ips[1].static_ip_address_id,
        aiven_static_ip.ips[2].static_ip_address_id,
        aiven_static_ip.ips[3].static_ip_address_id,
        aiven_static_ip.ips[4].static_ip_address_id,
        aiven_static_ip.ips[5].static_ip_address_id,
    ])

    pg_user_config {
        static_ips = true
    }
}
```

This leads to a rolling forward replacement of service nodes. The new nodes will use the static IP addresses.
