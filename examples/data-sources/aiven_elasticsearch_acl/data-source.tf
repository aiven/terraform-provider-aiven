data "aiven_elasticsearch_acl" "es-acls" {
    project = aiven_project.es-project.project
    service_name = aiven_elasticsearch.es.service_name
}

