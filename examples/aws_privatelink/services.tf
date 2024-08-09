# Kafka service
resource "aiven_kafka" "kafka-service" {
  project                 = data.aiven_project.pr1.project
  cloud_name              = "aws-eu-west-1"
  plan                    = "business-4"
  service_name            = "privatelink-kafka1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  project_vpc_id          = "AWS_PROJECT_VPC_ID"

  kafka_user_config {
    privatelink_access {
      kafka = true
    }

    public_access {
      kafka = true
    }

    kafka_authentication_methods {
      certificate = true
    }
  }
}

# AWS Privatelink service
resource "aiven_aws_privatelink" "aws_pl" {
  project      = data.aiven_project.pr1.project
  service_name = aiven_kafka.kafka-service.service_name
  principals   = [
    "arn:aws:iam::012345678901:role/my-privatelink-role",
  ]
}

output "aws_privatelink" {
  value = aiven_aws_privatelink.aws_pl
}

# After connecting to a VPC endpoint from your AWS account
# a new service component is available
data "aiven_service_component" "kafka_privatelink" {
  project                     = aiven_aws_privatelink.aws_pl.project
  service_name                = aiven_aws_privatelink.aws_pl.service_name
  component                   = "kafka"
  route                       = "privatelink"
  kafka_authentication_method = "certificate"
}

output "kafka_privatelink_component" {
  value = data.aiven_service_component.kafka_privatelink
}
