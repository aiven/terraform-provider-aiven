data "aiven_opensearch_acl_rule" "os_acl_rule" {
  project      = aiven_opensearch_acl_config.os_acls_config.project
  service_name = aiven_opensearch_acl_config.os_acls_config.service_name
  username     = "<USERNAME>"
  index        = "<INDEX>"
}
