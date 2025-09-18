# Project
resource "aiven_project" "kafka-con-project1" {
  project   = var.aiven_project
  parent_id = var.aiven_organization
}
