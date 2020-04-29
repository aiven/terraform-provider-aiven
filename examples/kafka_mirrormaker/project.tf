# Project
resource "aiven_project" "kafka-mm-project1" {
  project = "kafka-mm-project"
  card_id = var.aiven_card_id
}