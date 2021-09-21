# Project
resource "aiven_project" "os-project" {
  project = "os-project"
  card_id = var.aiven_card_id
}