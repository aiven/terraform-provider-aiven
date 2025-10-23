---
page_title: "Autoscale service capacity based on CPU usage"
---

# Autoscale service capacity based on CPU usage

Autoscaling for Aiven for Apache Kafka® Diskless Topics automatically adjusts your service capacity based on CPU usage. It helps maintain strong performance during traffic spikes by scaling up resources when the CPU load is consistently high. It also controls costs by scaling down when demand decreases. Autoscaling operates within the minimum and maximum plans that you define.

Autoscaling is available only for [Diskless Topics services deployed in Bring Your Own Cloud (BYOC) environments](https://aiven.io/docs/products/kafka/diskless/concepts/diskless-overview) using supported service plans.

~> **Limited availability**
Autoscaling for Diskless Topics is in the limited availability stage. To enable this feature, contact the [sales team](http://aiven.io/contact).

More information about how autoscaling for Diskless Topics works, and its benefits and limitations,
is available in the [Aiven documentation](https://aiven.io/docs/products/diskless/howto/enable-autoscaling).

## Enable service autoscaler

The `autoscaler_service` feature cannot be enabled during service creation using the Aiven Terraform Provider. To enable the autoscaler,
you first create the Kafka service with Diskless Topics enabled. After the service is created, you can use
the `aiven_service_integration_endpoint` and `aiven_service_integration` resources to enable the `autoscaler_service` feature.

1. Create a Kafka service. The following example creates a Kafka service with a `business-4` plan and enables Diskless Topics:

      ```hcl
      resource "aiven_kafka" "example_kafka" {
        project      = var.project_name
        cloud_name   = "google-europe-west1"
        plan         = "business-4"
        service_name = "example-kafka-diskless-staceys"
        kafka_user_config {
          kafka_diskless {
            enabled = true
          }
        }
      }

      ```

2. To create the service, run `terraform plan` and `terraform apply`.

3. To let the autoscaler to adjust the plan, remove the plan attribute from the Kafka resource after the service is created.

      ```hcl
      resource "aiven_kafka" "example_kafka" {
        project      = var.project_name
        cloud_name   = "google-europe-west1"
        service_name = "example-kafka-diskless-staceys"
        kafka_user_config {
          kafka_diskless {
            enabled = true
          }
        }
      }

      ```

4. To enable `autoscaler_service`, create a file with the autoscaler endpoint and service integration resources:

      ```hcl
      resource "aiven_service_integration_endpoint" "kafka_topic_autoscaler_endpoint" {
        project       = var.project_name
        endpoint_name = "topic-autoscaler"
        endpoint_type = "autoscaler_service"
      }

      resource "aiven_service_integration" "diskless-topics-autoscaler" {
        project             = var.project_name
        integration_type    = "autoscaler_service"
        source_service_name = aiven_kafka.example_kafka.service_name
        dest_endpoint_id    = aiven_service_integration_endpoint.kafka_topic_autoscaler_endpoint.id
        user_config = {
          autoscaling = {
            min_plan = "business-4"
            max_plan = "business-32"
          }
        }
      }

      ```

5. To preview the changes run `terraform plan`.

6. Apply the changes by running `terraform apply`.

You get an email notification whenever autoscaling scales the service up or down. You can also view the integration
in the [Aiven Console](https://console.aiven.io) by going to the service and clicking **Integrations**.

## Remove service autoscaler

To remove service autoscaling, remove the integration and choose a plan for your service.
The following example `main.tf` file has a Kafka service with the `autoscaler_service` integration:

```hcl
resource "aiven_kafka" "example_kafka" {
  project      = var.project_name
  cloud_name   = "google-europe-west1"
  service_name = "example-kafka-diskless-staceys"
  kafka_user_config {
    kafka_diskless {
      enabled = true
    }
  }
}

resource "aiven_service_integration_endpoint" "kafka_topic_autoscaler_endpoint" {
  project       = var.project_name
  endpoint_name = "topic-autoscaler"
  endpoint_type = "autoscaler_service"
}

resource "aiven_service_integration" "diskless-topics-autoscaler" {
  project             = var.project_name
  integration_type    = "autoscaler_service"
  source_service_name = aiven_kafka.example_kafka.service_name
  dest_endpoint_id    = aiven_service_integration_endpoint.kafka_topic_autoscaler_endpoint.id
  user_config = {
    autoscaling = {
      min_plan = "business-4"
      max_plan = "business-32"
    }
  }
}
```

1. To disable `autoscaler_service`, remove the `aiven_service_integration` resource.

      ```hcl
      resource "aiven_kafka" "example_kafka" {
        project      = var.project_name
        cloud_name   = "google-europe-west1"
        service_name = "example-kafka-diskless-staceys"
        kafka_user_config {
          kafka_diskless {
            enabled = true
          }
        }
      }

      resource "aiven_service_integration_endpoint" "kafka_topic_autoscaler_endpoint" {
        project       = var.project_name
        endpoint_name = "topic-autoscaler"
        endpoint_type = "autoscaler_service"
      }

      ```

2. Choose a `plan` for the Kafka service since the autoscaler will not be adjusting the plan.

      ```hcl
      resource "aiven_kafka" "example_kafka" {
        project      = var.project_name
        cloud_name   = "google-europe-west1"
        service_name = "example-kafka-diskless-staceys"
        plan         = "business-8"
        kafka_user_config {
          kafka_diskless {
            enabled = true
          }
        }
      }

      resource "aiven_service_integration_endpoint" "kafka_topic_autoscaler_endpoint" {
        project       = var.project_name
        endpoint_name = "topic-autoscaler"
        endpoint_type = "autoscaler_service"
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
