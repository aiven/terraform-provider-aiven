# Elasticsearch ACL Rule Resource

The Opensearch ACL Rule resource models a single ACL Rule for an Aiven Opensearch service.

## Example Usage

```hcl

resource "aiven_service_user" "es_user" {
    project = var.aiven_project_name
    service_name = aiven_opensearch.es_test.service_name
    username = "documentation-user-1"
}

resource "aiven_service_user" "es_user_2" {
    project = var.aiven_project_name
    service_name = aiven_opensearch.es_test.service_name
    username = "documentation-user-2"
}

resource "aiven_opensearch_acl_config" "es_acls_config" {
  project = var.aiven_project_name
  service_name = aiven_opensearch.es_test.service_name
  enabled = true
  extended_acl = false
}

locals {
  acl_rules = [
    {
      username = aiven_service_user.es_user.username
      index = "index2"
      permission = "readwrite"
    },
    {
      username = aiven_service_user.es_user.username
      index = "index3"
      permission = "read"
    },
    {
      username = aiven_service_user.es_user.username
      index = "index5"
      permission = "deny"
    },
    {
      username = aiven_service_user.es_user_2.username
      index = "index3"
      permission = "write"
    },
    {
      username = aiven_service_user.es_user_2.username
      index = "index7"
      permission = "readwrite"
    }
  ]
}

resource "aiven_opensearch_acl_rule" "es_acl_rule" {
  for_each = { for i, v in local.acl_rules:  i => v }

  project = aiven_opensearch_acl_config.es_acls_config.project
  service_name = aiven_opensearch_acl_config.es_acls_config.service_name
  username = each.value.username
  index = each.value.index
  permission = each.value.permission
}

```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the ACL belongs to. 
* `username` and `index` - (Required) define the username and index the ACL rule should apply to.

Changes to `project`, `service_name`, `username` or `index` will trigger recreation of the Opensearch ACL rule.

* `permission` - (Required) is the Opensearch permission, list of supported permissions: 
`deny`, `admin`, `read`, `readwrite`, `write`.
