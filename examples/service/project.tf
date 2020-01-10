# Project
resource "aiven_project" "project1" {
  project = "project1"
  card_id = var.aiven_card_id
}