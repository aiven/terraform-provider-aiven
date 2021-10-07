data "aiven_elasticsearch_acl_rule" "es_acl_rule" {
  project = aiven_elasticsearch_acl_config.es_acls_config.project
  service_name = aiven_elasticsearch_acl_config.es_acls_config.service_name
  username = "<USERNAME>"
  index = "<INDEX>"
}

