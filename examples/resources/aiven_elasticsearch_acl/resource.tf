resource "aiven_elasticsearch_acl" "es-acls" {
  project      = aiven_project.es-project.project
  service_name = aiven_service.es.service_name
  enabled      = true
  extended_acl = false
  acl {
    username = aiven_service_user.es-user.username
    rule {
      index      = "_*"
      permission = "admin"
    }

    rule {
      index      = "*"
      permission = "admin"
    }
  }

  acl {
    username = "avnadmin"
    rule {
      index      = "_*"
      permission = "read"
    }

    rule {
      index      = "*"
      permission = "read"
    }
  }
}
