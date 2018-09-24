variable "aiven_api_token" {}
variable "aiven_card_id" {}
variable "aiven_project_name" {}

# Initialize provider. No other config options than api_token
provider "aiven" {
	api_token = "${var.aiven_api_token}"
}

# Project
resource "aiven_project" "sample" {
    project = "${var.aiven_project_name}"
    card_id = "${var.aiven_card_id}"
}

# VPC for the project in AWS ap-south-1 region. Should also
# define peering connection(s) to make this useful.
resource "aiven_project_vpc" "vpc_aws_ap_south_1" {
	project = "${aiven_project.sample.project}"
	cloud_name = "aws-ap-south-1"
	network_cidr = "192.168.0.0/24"
}

# Kafka service
resource "aiven_service" "samplekafka" {
	project = "${aiven_project.sample.project}"
	cloud_name = "google-europe-west1"
	plan = "business-4"
	service_name = "samplekafka"
	service_type = "kafka"
	kafka_user_config {
		ip_filter = ["0.0.0.0/0"]
		kafka_connect = true
		kafka_rest = true
		kafka_version = "2.0"
		kafka {
			group_max_session_timeout_ms = 70000
			log_retention_bytes = 1000000000
		}
	}
}

# Topic for Kafka
resource "aiven_kafka_topic" "sample_topic" {
	project = "${aiven_project.sample.project}"
	service_name = "${aiven_service.samplekafka.service_name}"
	topic_name = "sample_topic"
	partitions = 3
	replication = 2
	retention_bytes = 1000000000
}

# User for Kafka
resource "aiven_service_user" "kafka_a" {
	project = "${aiven_project.sample.project}"
	service_name = "${aiven_service.samplekafka.service_name}"
	username = "kafka_a"
}

# ACL for Kafka
resource "aiven_kafka_acl" "sample_acl" {
	project = "${aiven_project.sample.project}"
        service_name = "${aiven_service.samplekafka.service_name}"
	username = "kafka_*"
	permission = "read"
	topic = "*"
}

# InfluxDB service
resource "aiven_service" "sampleinflux" {
	project = "${aiven_project.sample.project}"
	cloud_name = "google-europe-west1"
	plan = "startup-4"
	service_name = "sampleinflux"
	service_type = "influxdb"
	influxdb_user_config {
		ip_filter = ["0.0.0.0/0"]
	}
}

# Send metrics from Kafka to InfluxDB
resource "aiven_service_integration" "samplekafka_metrics" {
	project = "${aiven_project.sample.project}"
	integration_type = "metrics"
	source_service_name = "${aiven_service.samplekafka.service_name}"
	destination_service_name = "${aiven_service.sampleinflux.service_name}"
}

# PostreSQL service
resource "aiven_service" "samplepg" {
	project = "${aiven_project.sample.project}"
	cloud_name = "google-europe-west1"
	plan = "business-4"
	service_name = "samplepg"
	service_type = "pg"
	pg_user_config {
		ip_filter = ["0.0.0.0/0"]
		pg {
			idle_in_transaction_session_timeout = 900
		}
		pg_version = "10"
		pglookout {
			max_failover_replication_time_lag = 60
		}
	}
}

# Send metrics from PostgreSQL to InfluxDB
resource "aiven_service_integration" "samplepg_metrics" {
	project = "${aiven_project.sample.project}"
	integration_type = "metrics"
	source_service_name = "${aiven_service.samplepg.service_name}"
	destination_service_name = "${aiven_service.sampleinflux.service_name}"
}

# PostgreSQL database
resource "aiven_database" "sample_db" {
	project = "${aiven_project.sample.project}"
	service_name = "${aiven_service.samplepg.service_name}"
	database_name = "sample_db"
}

# PostgreSQL user
resource "aiven_service_user" "sample_user" {
	project = "${aiven_project.sample.project}"
	service_name = "${aiven_service.samplepg.service_name}"
	username = "sampleuser"
}

# Grafana service
resource "aiven_service" "samplegrafana" {
	project = "${aiven_project.sample.project}"
	cloud_name = "google-europe-west1"
	plan = "startup-4"
	service_name = "samplegrafana"
	service_type = "grafana"
	grafana_user_config {
		ip_filter = ["0.0.0.0/0"]
	}
}

# Dashboards for Kafka and PostgreSQL services
resource "aiven_service_integration" "samplegrafana_dashboards" {
	project = "${aiven_project.sample.project}"
	integration_type = "dashboard"
	source_service_name = "${aiven_service.samplegrafana.service_name}"
	destination_service_name = "${aiven_service.sampleinflux.service_name}"
}
