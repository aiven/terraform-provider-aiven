---
page_title: "Scale disk storage automatically"
---

# Scale disk storage automatically

[Disk autoscaler](https://aiven.io/docs/platform/howto/disk-autoscaler#disable-disk-autoscaler) automatically increases
the storage capacity of your Aiven service when it's running out of space. Disk autoscaler only increases storage, it doesn't scale down.

You can enable disk autoscaler when you create a service or on an existing service. To enable autoscaler, create
an autoscaler integration endpoint using the `aiven_service_integration_endpoint` resource and integrate
your service with it using the `aiven_service_integration` resource.

~> **Warning**
For services managed by Terraform, removing an autoscaler integration on services with `additional_disk_space` resets the service disk space to the service plan's disk size.
To retain the additional disk space set the service's `additional_disk_space` value manually. If the integration is managed by Terraform but not the service, the disk space is not reset.

## Enable disk autoscaler when creating a service

Follow [the disk autoscaler example](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/autoscaler_integration)
to learn how to to create an Aiven for PostgreSQL® service, autoscaler endpoint, and enable the autoscaler integration on the service.

## Enable disk autoscaler on an existing service

1. Create an autoscaler endpoint using the `aiven_service_integration_endpoint` resource. In the following example, an endpoint is created with a maximum total disk storage of 200GiB:

      ```hcl
      resource "aiven_service_integration_endpoint" "autoscaler_endpoint" {
        project       = var.aiven_project.example_project.project
        endpoint_name = "disk-autoscaler-200GiB"
        endpoint_type = "autoscaler"
        autoscaler_user_config {
          autoscaling {
            cap_gb = 200
            type   = "autoscale_disk"
          }
        }
      }
      ```

2. Enable the autoscaler integration on your service using the `aiven_service_integration` resource. For example:

      ```hcl
      resource "aiven_service_integration" "autoscaler_integration" {
        project                 = var.aiven_project.example_project.project
        integration_type        = "autoscaler"
        source_service_name     = "SERVICE_NAME"
        destination_endpoint_id = aiven_service_integration_endpoint.autoscaler_endpoint.id
      }
      ```

    Where `SERVICE_NAME` is the name of the service to scale.

3. To preview your configuration changes, run:

    ```bash
    terraform plan
    ```

4. To create the endpoint and enable the autoscaler, apply the configuration by running:

    ```bash
    terraform apply --auto-approve
     ```

5. If the service has `additional_disk_space` configured, remove the field. The additional storage will be adjusted by the autoscaler.
   
   -> After removing this field with the autoscaler enabled the output of `terraform plan` shows no changes.