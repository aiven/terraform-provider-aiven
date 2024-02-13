###
###  KAFKA
###
resource "aiven_kafka" "kafka-syd" {
  project      = var.aiven_api_project
  cloud_name   = "aws-ap-southeast-2"
  plan         = "business-4"
  service_name = "kafka-syd"

  kafka_user_config {
    kafka_version = "3.5"
    kafka_rest    = "true"
  }
}

resource "aiven_kafka" "kafka-use" {
  project      = var.aiven_api_project
  cloud_name   = "aws-us-east-1"
  plan         = "business-4"
  service_name = "kafka-us-east"

  kafka_user_config {
    kafka_version = "3.5"
    kafka_rest    = "true"
  }
}

resource "aiven_kafka" "kafka-usw" {
  project      = var.aiven_api_project
  cloud_name   = "aws-us-west-1"
  plan         = "business-4"
  service_name = "kafka-us-west"

  kafka_user_config {
    kafka_version = "3.5"
    kafka_rest    = "true"
  }
}

###
###  TOPICS
###
resource "aiven_kafka_topic" "topic-a-syd" {
  project      = var.aiven_api_project
  service_name = aiven_kafka.kafka-syd.service_name
  topic_name   = "topic-a"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_topic" "topic-a-use" {
  project      = var.aiven_api_project
  service_name = aiven_kafka.kafka-use.service_name
  topic_name   = "topic-a"
  partitions   = 3
  replication  = 2
}

resource "aiven_kafka_topic" "topic-a-usw" {
  project      = var.aiven_api_project
  service_name = aiven_kafka.kafka-usw.service_name
  topic_name   = "topic-a"
  partitions   = 3
  replication  = 2
}
