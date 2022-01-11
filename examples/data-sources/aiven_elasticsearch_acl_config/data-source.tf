data "aiven_elasticsearch_acl_config" "es-acl-config" {
  project      = aiven_project.es-project.project
  service_name = aiven_service.es.service_name
}
