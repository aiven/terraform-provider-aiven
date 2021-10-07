data "aiven_kafka_connect" "kc1" {
    project = data.aiven_project.pr1.project
    service_name = "my-kc1"
}

