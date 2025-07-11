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
See [Remove the autoscaler](#remove-the-autoscaler) for more.

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

5. If the service has `additional_disk_space` configured:
   - First apply the autoscaler configuration (steps 1-4)
   - Then in a separate Terraform run, remove the `additional_disk_space` field
   - The autoscaler will take over managing the disk space

   -> After removing this field with the autoscaler enabled the output of `terraform plan` shows no changes.

### Why does it take two runs?

The integration can't be created without a running service, and the `additional_disk_space` can't be removed without a running autoscaler. This produces a circular dependency.
Terraform doesn't share the changes graph with the Aiven Provider. All resources are managed individually, making it impossible to affect the plan of one resource depending on the plan of another.

### How can I adjust the disk space when the autoscaler is enabled?

Terraform is not designed for single-run operations. It would reset the disk size again and again on each run. That's why you can't set `additional_disk_space` while the autoscaler is running.

You can, however, adjust the disk space in the Aiven Console by [following the steps in the documentation](https://aiven.io/docs/platform/howto/add-storage-space).

## Remove the autoscaler on services with additional disk storage

When you remove the autoscaler integration, the service's disk space will be reset to the default size specified in the service plan. To prevent this automatic resizing:

1. Remove the `aiven_service_integration_endpoint` and `aiven_service_integration` autoscaler resources from your TF files.

   ~> Do not make any changes to the managed service. The `terraform plan` command must show no changes for it.

2. Apply the changes:

    ```bash
    terraform apply
    ```
3. Run `terraform plan` to confirm that autoscaler is no longer managing the service's disk storage. The output should show the `additional_disk_space` set to `0B` (the default value).

4. Set the `additional_disk_space` to the new value and apply your changes.

### Why can't I update the service in the same run?

When the autoscaler exists, the `additional_disk_space` can't be managed by users.
When it is destroyed during the `apply` command, the field is not set, hence it has no value (zero additional space). Terraform might return a `Provider produced inconsistent final plan` message or trigger the disk resize.
