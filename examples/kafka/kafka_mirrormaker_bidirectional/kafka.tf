###
###  KAFKA
###
resource "aiven_kafka" "kafka-ue1" {
  project      = var.aiven_api_project
  cloud_name   = "aws-us-east-1"
  plan         = "business-4"
  service_name = "kafka-ue1"

  kafka_user_config {
    kafka_version = "3.8"
    kafka_rest    = "true"
  }
}

resource "aiven_kafka" "kafka-uw2" {
  project      = var.aiven_api_project
  cloud_name   = "aws-us-west-2"
  plan         = "business-4"
  service_name = "kafka-uw2"

  kafka_user_config {
    kafka_version = "3.8"
    kafka_rest    = "true"
  }
}

###
###  TOPICS
###
resource "aiven_kafka_topic" "topic-a-ue1" {
  project      = var.aiven_api_project
  service_name = aiven_kafka.kafka-ue1.service_name
  topic_name   = "topic-a"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_topic" "topic-a-uw2" {
  project      = var.aiven_api_project
  service_name = aiven_kafka.kafka-uw2.service_name
  topic_name   = "topic-a"
  partitions   = 3
  replication  = 2
}
