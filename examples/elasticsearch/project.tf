# Project
resource "aiven_project" "es-project" {
  project = "es-project"
  card_id = var.aiven_card_id
}