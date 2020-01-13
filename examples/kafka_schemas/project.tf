# Project
resource "aiven_project" "kafka-schemas-project1" {
  project = "kafka-schemas-project1"
  card_id = var.aiven_card_id
}