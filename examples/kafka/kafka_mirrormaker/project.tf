# Project
resource "aiven_project" "kafka-mm-project1" {
  project   = var.aiven_project
  parent_id = var.aiven_organization
}
