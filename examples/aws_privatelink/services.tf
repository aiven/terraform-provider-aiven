# Kafka service
resource "aiven_kafka" "kafka-service" {
  project = data.aiven_project.pr1.project
  cloud_name = "aws-eu-west-1"
  plan = "business-4"
  service_name = "privatelink-kafka1"
  maintenance_window_dow = "monday"
  maintenance_window_time = "10:00:00"
    project_vpc_id = "YOUR-AWS-PROJECT-VPC-ID"
}

# AWS Privatelink service
resource "aiven_aws_privatelink" "aws_pl" {
  project = data.aiven_project.pr1.project
  service_name = aiven_kafka.kafka-service.service_name
  principals = [
    "arn:aws:iam::012345678901:role/my-privatelink-role"]
}

data "aiven_aws_privatelink" "pl" {
  project = aiven_aws_privatelink.aws_pl.project
  service_name = aiven_aws_privatelink.aws_pl.service_name
}

output "aws_privatelink" {
  value = aiven_aws_privatelink.aws_pl
}