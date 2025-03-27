locals {
  acl_rules = [
    {
      username   = aiven_opensearch_user.os_user.username
      index      = "_*"
      permission = "admin"
    },
    {
      username   = aiven_opensearch_user.os_user.username
      index      = "*"
      permission = "admin"
    },

    # avnadmin is a default user created by the Aiven for OpenSearch service
    # It has admin ACL by default

    {
      username   = "avnadmin"
      index      = "_*"
      permission = "read"
    },
    {
      username   = "avnadmin"
      index      = "*"
      permission = "read"
    },
  ]
}

resource "aiven_opensearch_acl_config" "acls" {
  project      = data.aiven_project.main.project
  service_name = aiven_opensearch.example_opensearch.service_name
  enabled      = true
  extended_acl = false
}

resource "aiven_opensearch_acl_rule" "os_acl_rule" {
  for_each = { for i, v in local.acl_rules : i => v }

  project      = aiven_opensearch_acl_config.acls.project
  service_name = aiven_opensearch_acl_config.acls.service_name

  username   = each.value.username
  index      = each.value.index
  permission = each.value.permission
}
