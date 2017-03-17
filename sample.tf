variable "aiven_email" {}
variable "aiven_password" {}
variable "aiven_card_id" {}

provider "aiven" {
	email       = "${var.aiven_email}"
	password    = "${var.aiven_password}"
}

resource "aiven_project" "sample" {
    project = "sample"
    card_id = "${var.aiven_card_id}"
    cloud = "google-europe-west1"
}

resource "aiven_service" "postgresql" {
	project = "${aiven_project.sample.project}"
	group_name = "test"
    cloud = "google-europe-west1"
	plan = "hobbyist"
	service_name = "test-postgresql"
	service_type = "pg"
}

resource "aiven_database" "postgresql" {
	project = "${aiven_service.postgresql.project}"
	service_name = "${aiven_service.postgresql.service_name}"
	database = "coda"
}

resource "aiven_service_user" "postgresql" {
	project = "${aiven_service.postgresql.project}"
	service_name = "${aiven_database.postgresql.service_name}"
	username = "codabox"
}
