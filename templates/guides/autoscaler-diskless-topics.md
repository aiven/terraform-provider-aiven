---
page_title: "Autoscale service capacity based on CPU usage"
---

# Autoscale service capacity based on CPU usage

~> **Limited availability**
Autoscaling Inkless clusters is in the limited availability stage. To enable this feature, contact the [Aiven support team](mailto:support@aiven.io).

Autoscaling automatically adjusts your service capacity based on CPU usage. It helps maintain strong performance during traffic spikes by scaling up resources when the CPU load is consistently high. It also controls costs by scaling down when demand decreases. Autoscaling operates within the minimum and maximum plans that you define.

## Prerequisites

Autoscaling is only available for Inkless services. Inkless services are `aiven_kafka` services with the following:
  -  A plan with the `-inkless` suffix
  -  Tiered storage enabled
  -  Kafka version 4.0 or higher

For more information on key considerations and limitations, see the [Aiven documentation](https://aiven.io/docs/products/diskless/howto/enable-autoscaling).

## Enable service autoscaler

The `autoscaler_service` feature cannot be enabled during service creation using the Aiven Terraform Provider. To enable the autoscaler,
first create the Inkless service. After the service is created, you can use
the `aiven_service_integration_endpoint` and `aiven_service_integration` resources to enable the `autoscaler_service` feature.

1. Create an Inkless service. The following is an example Inkless service with the `business-8-inkless` plan:

      ```hcl
      resource "aiven_kafka" "example_inkless" {
        project      = var.project_name
        cloud_name   = var.cloud_name
        plan         = "business-8-inkless"
        service_name = "example-inkless-autoscaler"
        kafka_user_config {
          kafka_version = "4.0"
          kafka_diskless {
            enabled = true
          }
          tiered_storage {
            enabled = true
          }
        }
      }
      ```

2. To create the service, run `terraform plan` and `terraform apply`.

   -> An Aiven for PostgreSQL service is also automatically created for Inkless services. You don't need to manage this service.

3. To let the autoscaler adjust the plan, remove the plan attribute from the `aiven_kafka` resource after the service is created.

      ```hcl
      resource "aiven_kafka" "example_inkless" {
        project      = var.project_name
        cloud_name   = var.cloud_name
        service_name = "example-inkless-autoscaler"
        kafka_user_config {
          kafka_version = "4.0"
          kafka_diskless {
            enabled = true
          }
          tiered_storage {
            enabled = true
          }
        }
      }
      ```

4. To enable `autoscaler_service`, create a file with the autoscaler endpoint and service integration resources:

      ```hcl
      resource "aiven_service_integration_endpoint" "topic_autoscaler_endpoint" {
        project       = var.project_name
        endpoint_name = "topic-autoscaler"
        endpoint_type = "autoscaler_service"
      }

      resource "aiven_service_integration" "diskless-topics-autoscaler" {
        project             = var.project_name
        integration_type    = "autoscaler_service"
        source_service_name = aiven_kafka.example_inkless.service_name
        dest_endpoint_id    = aiven_service_integration_endpoint.topic_autoscaler_endpoint.id
        user_config = {
          autoscaling = {
            min_plan = "business-8-inkless"
            max_plan = "business-32-inkless"
          }
        }
      }
      ```

5. To preview the changes, run: 

     ```hcl
     terraform plan
     ```

6. Apply the changes by running: 

     ```hcl
     terraform apply
     ```

You get an email notification whenever autoscaling scales the service up or down.

## View the service plan

To view the current service plan, you can add an output block to your Terraform configuration, for example:

```hcl
output "example_inkless_service_plan" {
  value = aiven_kafka.example_inkless.plan
}
```

To view the service plan, run:

```bash
terraform output example_inkless_service_plan
```

The output will be similar to:

```bash
example_inkless_service_plan = "business-16-inkless"
```

To view the service plan in the [Aiven Console](https://console.aiven.io), go to the service overview.

## Remove service autoscaler

To remove service autoscaling, remove the integration and choose a plan for your service.
The following example file has an Inkless service with the `autoscaler_service` integration:

```hcl
resource "aiven_kafka" "example_inkless" {
  project      = var.project_name
  cloud_name   = var.cloud_name
  service_name = "example-inkless-autoscaler"
  kafka_user_config {
    kafka_version = "4.0"
    kafka_diskless {
      enabled = true
    }
    tiered_storage {
      enabled = true
    }
  }
}

resource "aiven_service_integration_endpoint" "topic_autoscaler_endpoint" {
  project       = var.project_name
  endpoint_name = "topic-autoscaler"
  endpoint_type = "autoscaler_service"
}

resource "aiven_service_integration" "diskless-topics-autoscaler" {
  project             = var.project_name
  integration_type    = "autoscaler_service"
  source_service_name = aiven_kafka.example_inkless.service_name
  dest_endpoint_id    = aiven_service_integration_endpoint.topic_autoscaler_endpoint.id
  user_config = {
    autoscaling = {
      min_plan = "business-8-inkless"
      max_plan = "business-32-inkless"
    }
  }
}
```

1. To disable `autoscaler_service`, remove the `aiven_service_integration` and `aiven_service_integration_endpoint` resources.

      ```hcl
      resource "aiven_kafka" "example_inkless" {
        project      = var.project_name
        cloud_name   = var.cloud_name
        service_name = "example-inkless-autoscaler"
        kafka_user_config {
          kafka_version = "4.0"
          kafka_diskless {
            enabled = true
          }
          tiered_storage {
            enabled = true
          }
        }
      }
      ```

2. Choose a `plan` for the Inkless service since the autoscaler will not be adjusting the plan.

      ```hcl
      resource "aiven_kafka" "example_inkless" {
        project      = var.project_name
        cloud_name   = var.cloud_name
        plan         = "business-8" # Add a service plan
        service_name = "example-inkless-autoscaler"
        kafka_user_config {
          kafka_version = "4.0"
          kafka_diskless {
            enabled = true
          }
          tiered_storage {
            enabled = true
          }
        }
      }
      ```

3. To preview your changes, run:

     ```hcl
     terraform plan
     ```

4. To apply your changes, run:

     ```hcl
     terraform apply
     ```
