###
###  MIRRORMAKER
###
resource "aiven_kafka_mirrormaker" "mm" {
  project      = var.aiven_api_project
  cloud_name   = "aws-ap-southeast-1"
  plan         = "startup-4"
  service_name = "mm"

  kafka_mirrormaker_user_config {
    ip_filter = [
      "0.0.0.0/0"
    ]

    kafka_mirrormaker {
      refresh_groups_interval_seconds = 600
      refresh_topics_enabled          = true
      refresh_topics_interval_seconds = 600
    }
  }
}

###
###  CLUSTER ALIASES
###
resource "aiven_service_integration" "mm-int-syd" {
  project                  = var.aiven_api_project
  integration_type         = "kafka_mirrormaker"
  source_service_name      = aiven_kafka.kafka-syd.service_name
  destination_service_name = aiven_kafka_mirrormaker.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = aiven_kafka.kafka-syd.service_name
  }
}

resource "aiven_service_integration" "mm-int-use" {
  project                  = var.aiven_api_project
  integration_type         = "kafka_mirrormaker"
  source_service_name      = aiven_kafka.kafka-use.service_name
  destination_service_name = aiven_kafka_mirrormaker.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = aiven_kafka.kafka-use.service_name
  }
}

resource "aiven_service_integration" "mm-int-usw" {
  project                  = var.aiven_api_project
  integration_type         = "kafka_mirrormaker"
  source_service_name      = aiven_kafka.kafka-usw.service_name
  destination_service_name = aiven_kafka_mirrormaker.mm.service_name

  kafka_mirrormaker_user_config {
    cluster_alias = aiven_kafka.kafka-usw.service_name
  }
}

###
###  REPLICATION FLOWS
###
resource "aiven_mirrormaker_replication_flow" "f1" {
  project        = var.aiven_api_project
  service_name   = aiven_kafka_mirrormaker.mm.service_name
  source_cluster = aiven_kafka.kafka-syd.service_name
  target_cluster = aiven_kafka.kafka-use.service_name
  enable         = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]
}

resource "aiven_mirrormaker_replication_flow" "f2" {
  project        = var.aiven_api_project
  service_name   = aiven_kafka_mirrormaker.mm.service_name
  source_cluster = aiven_kafka.kafka-use.service_name
  target_cluster = aiven_kafka.kafka-usw.service_name
  enable         = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]
}

resource "aiven_mirrormaker_replication_flow" "f3" {
  project        = var.aiven_api_project
  service_name   = aiven_kafka_mirrormaker.mm.service_name
  source_cluster = aiven_kafka.kafka-usw.service_name
  target_cluster = aiven_kafka.kafka-syd.service_name
  enable         = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]
}
