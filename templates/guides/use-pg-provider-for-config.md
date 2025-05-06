---
page_title: "Use PostgreSQL® Provider to configure Aiven for PostgreSQL® services"
---

# Use PostgreSQL® Provider to configure Aiven for PostgreSQL® services

Use the [PostgreSQL Provider for Terraform](https://registry.terraform.io/providers/cyrilgdn/postgresql/latest/docs) to configure settings such as default privileges, publication, or to reuse a submodule between different vendors.

You can create an Aiven for PostgreSQL® service with the Aiven Terraform Provider and configure it with the PostgreSQL Provider.

1.  Add the PostgreSQL Provider and Aiven Terraform Provider to the
    `required_providers` block:

    ```hcl
    terraform {
      required_providers {
        aiven = {
          source  = "aiven/aiven"
          version = ">=4.0.0, < 5.0.0"
        }
        postgresql = {
          source  = "cyrilgdn/postgresql"
          version = "1.25.0"
        }
      }
    }
    ```

2.  Set the service connection attributes in the `provider` block:

    ```hcl
    # Aiven service
    resource "aiven_pg" "example_pg" {
      project                 = data.aiven_project.example_project.project
      cloud_name              = "google-asia-southeast1"
      plan                    = "business-8"
      service_name            = "example-pg-service"
    }

    # Configure the PostgreSQL Provider by referencing the Aiven service resource
    provider "postgresql" {
      host            = aiven_pg.example_pg.service_host
      port            = aiven_pg.example_pg.service_port
      database        = aiven_pg.example_pg.pg.dbname
      username        = aiven_pg.example_pg.service_username
      password        = aiven_pg.example_pg.service_password
      superuser       = false
      sslmode         = "require"
      connect_timeout = 15
    }
    ```

3.  Use the
    [PostgreSQL Provider resources](https://registry.terraform.io/providers/cyrilgdn/postgresql/latest/docs)
    to configure your service.
