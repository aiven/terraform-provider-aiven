resource "aiven_kafka_acl" "mytestacl" {
    project = aiven_project.myproject.project
    service_name = aiven_kafka.myservice.service_name
    topic = "<TOPIC_NAME_PATTERN>"
    permission = "admin"
    username = "<USERNAME_PATTERN>"
}
