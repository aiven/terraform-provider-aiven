# Project
resource "aiven_project" "kafka-con-project1" {
  project = "kafka-con-project1"
  card_id = var.aiven_card_id
}