data "aiven_opensearch_acl_config" "os-acl-config" {
  project      = aiven_project.os-project.project
  service_name = aiven_service.os.service_name
}

