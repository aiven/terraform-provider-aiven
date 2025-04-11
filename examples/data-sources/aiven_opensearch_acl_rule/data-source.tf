data "aiven_opensearch_acl_rule" "os_acl_rule" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_opensearch.example_opensearch.service_name
  username     = "documentation-user-1"
  index        = "index5"
}

