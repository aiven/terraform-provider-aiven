data "aiven_opensearch_acl_config" "os-acl-config" {
  project      = aiven_project.os-project.project
  service_name = aiven_opensearch.os.service_name
}
