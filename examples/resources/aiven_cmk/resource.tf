resource "aiven_cmk" "example_user" {
  project = var.aiven_project_name
  resource = var.cmk_resource
  cmk_provider = "gcp"
  default_cmk = false
}
