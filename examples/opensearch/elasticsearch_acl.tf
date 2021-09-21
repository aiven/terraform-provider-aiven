locals {
  acl_rules = [
    {
      username = aiven_service_user.es-user.username
      index = "_*"
      permission = "admin"
    },
    {
      username = aiven_service_user.es-user.username
      index = "*"
      permission = "admin"
    },

    # avnadmin is a default user created by Aivan for Opensearch service, and it has admin ACL by default

    {
      username = "avnadmin"
      index = "_*"
      permission = "read"
    },
    {
      username = "avnadmin"
      index = "*"
      permission = "read"
    },
  ]
}

resource "aiven_opensearch_acl_config" "os_acls_config" {
  project = aiven_opensearch.os.project
  service_name = aiven_opensearch.os.service_name
  enabled = true
  extended_acl = false
}

resource "aiven_opensearch_acl_rule" "os_acl_rule" {
  for_each = {for i, v in local.acl_rules:  i => v}

  project = aiven_opensearch_acl_config.os_acls_config.project
  service_name = aiven_opensearch_acl_config.os_acls_config.service_name

  username = each.value.username
  index = each.value.index
  permission = each.value.permission
}
